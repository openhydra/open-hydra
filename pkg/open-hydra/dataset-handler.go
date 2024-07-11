package openhydra

import (
	"archive/zip"
	"bytes"
	stdErr "errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xDatasetV1 "open-hydra/pkg/apis/open-hydra-api/dataset/core/v1"

	"github.com/emicklei/go-restful/v3"
	simpleChinese "golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

/*
 * Dataset routes
 * for posted data is a zip file we will extract it to a temp dir -> config.PublicDatasetBasePath/{dataset-name}
 * if file is already exist, we will return 409 for post request
 * for put request, we will overwrite the file
 * may not working with large file because of http request size limit
 */
func (builder *OpenHydraRouteBuilder) AddDatasetListRoute() {
	path := "/" + DatasetPath
	builder.addPathAuthorization(path, http.MethodGet, 1)
	builder.RootWS.Route(builder.RootWS.GET(path).Operation("listDataset").To(builder.DatasetListRouteHandler).
		Returns(http.StatusInternalServerError, "internal server error", "").
		Returns(http.StatusForbidden, "forbidden", "").
		Returns(http.StatusUnauthorized, "unauthorized", "").
		Returns(http.StatusOK, "OK", xDatasetV1.DatasetList{}))
}

func (builder *OpenHydraRouteBuilder) DatasetListRouteHandler(request *restful.Request, response *restful.Response) {
	datasetList, err := builder.Database.ListDatasets()
	if err != nil {
		writeAPIStatusError(response, err)
		return
	}
	datasetList.Kind = "List"
	datasetList.APIVersion = "v1"
	response.WriteEntity(datasetList)
}

func (builder *OpenHydraRouteBuilder) AddDatasetGetRoute() {
	path := "/" + DatasetPath + "/{dataset-name}"
	builder.addPathAuthorization(path, http.MethodGet, 1)
	builder.RootWS.Route(builder.RootWS.GET(path).Operation("getDataset").To(builder.DatasetGetRouteHandler).
		Returns(http.StatusNotFound, "not found", "").
		Returns(http.StatusInternalServerError, "internal server error", "").
		Returns(http.StatusUnauthorized, "unauthorized", "").
		Returns(http.StatusOK, "OK", xDatasetV1.Dataset{}))
}

func (builder *OpenHydraRouteBuilder) DatasetGetRouteHandler(request *restful.Request, response *restful.Response) {
	datasetName := request.PathParameter("dataset-name")
	dataset, err := builder.Database.GetDataset(datasetName)
	if err != nil {
		writeAPIStatusError(response, err)
		return
	}
	response.WriteEntity(dataset)
}

func (builder *OpenHydraRouteBuilder) AddDatasetCreateRoute() {
	path := "/" + DatasetPath
	builder.addPathAuthorization(path, http.MethodPost, 1)
	builder.RootWS.Route(builder.RootWS.POST(path).Operation("createDataset").To(builder.DatasetCreateRouteHandler).
		Returns(http.StatusCreated, "created", "").
		Returns(http.StatusBadRequest, "bad request", "").
		Returns(http.StatusInternalServerError, "internal server error", "").
		Returns(http.StatusForbidden, "forbidden", "").
		Returns(http.StatusUnauthorized, "unauthorized", "").
		Consumes("multipart/form-data"))
}

func (builder *OpenHydraRouteBuilder) DatasetCreateRouteHandler(request *restful.Request, response *restful.Response) {
	err := request.Request.ParseMultipartForm(builder.Config.PublicDatasetMaxSize)
	if err != nil {
		writeHttpResponseAndLogError(response, http.StatusBadRequest,
			fmt.Sprintf("Failed to parse multipart form: %v", err))
		return
	}
	description := request.Request.PostFormValue("description")
	name := request.Request.PostFormValue("name")
	if name == "" {
		writeHttpResponseAndLogError(response, http.StatusBadRequest, "Dataset name is empty")
		return
	}

	file, fileHeader, err := request.Request.FormFile("file")
	if err != nil {
		writeHttpResponseAndLogError(response, http.StatusBadRequest,
			err.Error())
		return
	}
	defer file.Close()
	if strings.ToLower(filepath.Ext(fileHeader.Filename)) != ".zip" {
		writeHttpResponseAndLogError(response, http.StatusBadRequest,
			"Only zip file is supported")
		return
	}
	_, err = builder.Database.GetDataset("name")
	if err == nil {
		writeHttpResponseAndLogError(response, http.StatusConflict,
			fmt.Sprintf("Dataset %s already exists", name))
		return
	}
	if !errors.IsNotFound(err) {
		writeAPIStatusError(response, err)
		return
	}

	datasetPath, err := filepath.Abs(filepath.Join(builder.Config.PublicDatasetBasePath, name))
	if err != nil {
		writeHttpResponseAndLogError(response, http.StatusBadRequest,
			fmt.Sprintf("Failed to get dataset path: %v", err.Error()))
		return
	}
	err = unzipTo(file, fileHeader.Size, datasetPath)
	if err != nil {
		writeAPIStatusError(response, err)
		return
	}
	dataset := &xDatasetV1.Dataset{
		ObjectMeta: metaV1.ObjectMeta{
			Name: name,
		},
		Spec: xDatasetV1.DatasetSpec{
			Description: description,
		},
	}

	err = builder.Database.CreateDataset(dataset)
	if err != nil {
		writeHttpResponseAndLogError(response, http.StatusInternalServerError,
			fmt.Sprintf("Failed create dateset: %v", err.Error()))
		return
	}
	response.WriteHeaderAndEntity(http.StatusCreated, dataset)
}

func (builder *OpenHydraRouteBuilder) AddDatasetUpdateRoute() {
	path := "/" + DatasetPath + "/{dataset-name}"
	builder.addPathAuthorization(path, http.MethodPut, 1)
	builder.RootWS.Route(builder.RootWS.PUT(path).Operation("getUpdateDataset").To(builder.DatasetUpdateRouteHandler).
		Returns(http.StatusNotFound, "not found", "").
		Returns(http.StatusBadRequest, "bad request", "").
		Returns(http.StatusInternalServerError, "internal server error", "").
		Returns(http.StatusForbidden, "forbidden", "").
		Returns(http.StatusUnauthorized, "unauthorized", "").
		Consumes("multipart/form-data"))
}

func (builder *OpenHydraRouteBuilder) DatasetUpdateRouteHandler(request *restful.Request, response *restful.Response) {
	datasetName := request.PathParameter("dataset-name")
	err := request.Request.ParseMultipartForm(builder.Config.PublicDatasetMaxSize)
	if err != nil {
		writeHttpResponseAndLogError(response, http.StatusBadRequest,
			fmt.Sprintf("Failed to parse multipart form: %v", err))
		return
	}
	file, fileHeader, err := request.Request.FormFile("file")
	if err != nil {
		if stdErr.Is(err, http.ErrMissingFile) {
			response.WriteHeader(http.StatusOK)
			return
		}
		writeHttpResponseAndLogError(response, http.StatusBadRequest,
			err.Error())
		return
	}
	defer file.Close()
	if strings.ToLower(filepath.Ext(fileHeader.Filename)) != ".zip" {
		writeHttpResponseAndLogError(response, http.StatusBadRequest,
			"Only zip file is supported")
		return
	}

	description := request.Request.PostFormValue("description")
	dataset, err := builder.Database.GetDataset(datasetName)
	if err != nil {
		writeAPIStatusError(response, err)
		return
	}
	dataset.Spec.Description = description
	err = builder.Database.UpdateDataset(dataset)
	if err != nil {
		writeAPIStatusError(response, err)
		return
	}

	// update dataset file

	datasetPath, err := filepath.Abs(filepath.Join(builder.Config.PublicDatasetBasePath, dataset.Name))
	if err != nil {
		writeHttpResponseAndLogError(response, http.StatusBadRequest,
			fmt.Sprintf("Failed to get dataset path: %v", err.Error()))
		return
	}
	err = unzipTo(file, fileHeader.Size, datasetPath)
	if err != nil {
		writeAPIStatusError(response, err)
		return
	}
	response.WriteHeader(http.StatusOK)
}

func (builder *OpenHydraRouteBuilder) AddDatasetDeleteRoute() {
	path := "/" + DatasetPath + "/{dataset-name}"
	builder.addPathAuthorization(path, http.MethodDelete, 1)
	builder.RootWS.Route(builder.RootWS.DELETE(path).Operation("deleteDataset").To(builder.DatasetDeleteRouteHandler).
		Returns(http.StatusNotFound, "not found", "").
		Returns(http.StatusInternalServerError, "internal server error", "").
		Returns(http.StatusForbidden, "forbidden", "").
		Returns(http.StatusUnauthorized, "unauthorized", ""))
}

func (builder *OpenHydraRouteBuilder) DatasetDeleteRouteHandler(request *restful.Request, response *restful.Response) {
	datasetName := request.PathParameter("dataset-name")
	err := builder.Database.DeleteDataset(datasetName)
	if err != nil {
		writeAPIStatusError(response, err)
		return
	}
	datasetPath, err := filepath.Abs(filepath.Join(builder.Config.PublicDatasetBasePath, datasetName))
	if err != nil {
		writeHttpResponseAndLogError(response, http.StatusInternalServerError,
			fmt.Sprintf("Failed to get dataset path: %v", err.Error()))
	}
	err = os.RemoveAll(datasetPath)
	if err != nil {
		writeHttpResponseAndLogError(response, http.StatusInternalServerError,
			fmt.Sprintf("Failed to remove dataset file: %v", err.Error()))
		return
	}
	response.WriteHeader(http.StatusOK)
}

func unzipTo(src io.ReaderAt, size int64, destPath string) error {
	reader, err := zip.NewReader(src, size)
	if err != nil {
		return errors.NewBadRequest(fmt.Sprintf("Invalid zip file: %v", err.Error()))
	}
	tempPath := fmt.Sprintf("%s-%s", destPath, strconv.FormatInt(time.Now().UnixNano(), 36))
	err = os.MkdirAll(tempPath, os.ModePerm)
	if err != nil {
		return errors.NewInternalError(err)
	}
	defer os.RemoveAll(tempPath)
	for _, file := range reader.File {
		// for chinese file name that compressed on windows might cause name is garbled
		// to fix this we have to always convert file name to format B18030
		decodedFileName, err := decodeGB18030ToString([]byte(file.Name))
		if err != nil {
			return errors.NewBadRequest(fmt.Sprintf("Failed decoded file name '%s' to format GB18030 due to error: %v", file.Name, err.Error()))
		}

		err = writeZipFile(filepath.Join(tempPath, decodedFileName), file)
		if err != nil {
			return errors.NewInternalError(err)
		}
	}
	_ = os.RemoveAll(destPath)
	err = os.Rename(tempPath, destPath)
	if err != nil {
		return errors.NewInternalError(err)
	}
	return nil
}

func writeZipFile(destPah string, zipFile *zip.File) error {
	if zipFile.FileInfo().IsDir() {
		// for those dir is separate from file
		// if is a dir only create it is all we need to do
		return os.MkdirAll(destPah, os.ModePerm)
	}

	// for dir is not separate from file
	// we have to create dir first
	dir, _ := filepath.Split(destPah)
	_ = os.MkdirAll(dir, os.ModePerm)
	fh, err := os.Create(destPah)
	if err != nil {
		return err
	}
	defer fh.Close()
	zh, err := zipFile.Open()
	if err != nil {
		return err
	}
	defer zh.Close()
	_, err = io.Copy(fh, zh)
	return err
}

func decodeGB18030ToString(data []byte) (string, error) {
	reader := transform.NewReader(bytes.NewReader(data), simpleChinese.GB18030.NewDecoder())
	decoded, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}
