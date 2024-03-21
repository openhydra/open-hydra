package app

import (
	"context"
	"fmt"
	"log/slog"
	"open-hydra/cmd/open-hydra-server/app/config"
	"open-hydra/cmd/open-hydra-server/app/option"
	"open-hydra/pkg/apiserver"
	"strings"

	"github.com/common-nighthawk/go-figure"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

func NewCommand(version string) *cobra.Command {
	options := &option.Options{
		OpenHydraServerOption: option.NewDefaultOpenHydraServerOption(),
		ApiServerOption:       option.NewDefaultApiServerOption(),
	}
	cmd := &cobra.Command{
		Use:     "open-hydra-server",
		Example: figure.NewColorFigure("OpenHydra", "", "green", true).String(),
	}

	runCmd := &cobra.Command{
		Use:     "run",
		Short:   "Launch a open-hydra-server api with configuration file",
		Long:    "run subcommand will launch a open-hydra-serverr",
		Example: "open-hydra-server run",
		RunE:    run(options, signals.SetupSignalHandler().Done()),
	}

	// add version api

	verCmd := &cobra.Command{
		Use:     "version",
		Short:   "Print version and exit",
		Long:    "version subcommand will print version and exit",
		Example: "open-hydra-server version",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Println("version:", version)
		},
	}

	options.BindFlags(runCmd.Flags())
	cmd.AddCommand(runCmd, verCmd)

	return cmd
}

func run(options *option.Options, stopCh <-chan struct{}) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func() {
			<-stopCh
			slog.Info("Received termination, signaling shutdown.")
			cancel()
		}()

		openHydraConfig, err := config.LoadConfig(options.OpenHydraServerOption.ConfigFile, options.OpenHydraServerOption.KubeConfigFile)
		if err != nil {
			slog.Error(fmt.Sprintf("Failed to load open-hydra-server config file: %v", err))
			return err
		}

		errMsg := checkConfig(openHydraConfig)
		if len(errMsg) > 0 {
			slog.Error(fmt.Sprintf("Failed to check open-hydra-server config file: %v", strings.Join(errMsg, ",")))
		}

		onLeaderRun := func(ctx context.Context) {
			err := apiserver.RunApiServer(options, openHydraConfig, ctx.Done())
			if err != nil {
				panic(err)
			}
		}
		if openHydraConfig.LeaderElection.LeaderElect {
			kubeClient, err := kubernetes.NewForConfig(openHydraConfig.KubeConfig)
			if err != nil {
				panic(err)
			}
			id := uuid.New().String()
			lock := &resourcelock.LeaseLock{
				LeaseMeta: metaV1.ObjectMeta{
					Name:      openHydraConfig.LeaderElection.ResourceName,
					Namespace: openHydraConfig.LeaderElection.ResourceNamespace,
				},
				Client: kubeClient.CoordinationV1(),
				LockConfig: resourcelock.ResourceLockConfig{
					Identity: id,
				},
			}

			// start the leader election loop
			leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
				Lock:            lock,
				ReleaseOnCancel: true,
				LeaseDuration:   openHydraConfig.LeaderElection.LeaseDuration,
				RenewDeadline:   openHydraConfig.LeaderElection.RenewDeadline,
				RetryPeriod:     openHydraConfig.LeaderElection.RetryPeriod,
				Callbacks: leaderelection.LeaderCallbacks{
					OnStartedLeading: onLeaderRun,
					OnStoppedLeading: func() {
						slog.Info(fmt.Sprintf("leader lost: %s", id))
					},
					OnNewLeader: func(identity string) {
						// we're notified when new leader elected
						if identity == id {
							// I just got the lock
							return
						}
						slog.Info(fmt.Sprintf("new leader elected: %s", identity))
					},
				},
			})

			return nil
		} else {
			onLeaderRun(ctx)
		}
		return nil
	}
}

func checkConfig(config *config.OpenHydraServerConfig) []string {
	var errMsg []string
	err := checkDBConfig(config)
	if err != nil {
		errMsg = append(errMsg, err.Error())
	}
	return errMsg
}

func checkDBConfig(config *config.OpenHydraServerConfig) error {
	// add a comment for testing
	if config.MySqlConfig == nil || config.EtcdConfig == nil {
		return fmt.Errorf("both mysql and etcd config are nil, at least one of them should be set")
	}

	oneDBConfigIsOk := false
	if config.MySqlConfig != nil {
		if config.MySqlConfig.Address == "" {
			return fmt.Errorf("mysql address is empty")
		}

		if config.MySqlConfig.Port == 0 {
			return fmt.Errorf("mysql port is empty")
		}

		if config.MySqlConfig.Username == "" {
			return fmt.Errorf("mysql username is empty")
		}

		if config.MySqlConfig.Password == "" {
			return fmt.Errorf("mysql password is empty")
		}
		config.DBType = "mysql"
		oneDBConfigIsOk = true
	}

	if config.EtcdConfig != nil && !oneDBConfigIsOk {
		if len(config.EtcdConfig.Endpoints) == 0 {
			return fmt.Errorf("etcd endpoints is empty")
		}

		if strings.Contains(config.EtcdConfig.Endpoints[0], "https") {
			if config.EtcdConfig.CAFile == "" {
				return fmt.Errorf("etcd ca file is empty")
			}

			if config.EtcdConfig.CertFile == "" {
				return fmt.Errorf("etcd cert file is empty")
			}

			if config.EtcdConfig.KeyFile == "" {
				return fmt.Errorf("etcd key file is empty")
			}
		}
		config.DBType = "etcd"
		oneDBConfigIsOk = true
	}

	if !oneDBConfigIsOk {
		return fmt.Errorf("both mysql and etcd config are empty, at least one of them should be set")
	}

	return nil
}
