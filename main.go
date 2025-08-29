package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"sync"
	"tssc/internal/appdata"
	"tssc/internal/proxy/dns"
	"tssc/internal/proxy/ipv6"
	"tssc/internal/proxy/outline_device"
	"tssc/internal/proxy/routing"
	tundevice "tssc/internal/proxy/tun_device"

	"github.com/urfave/cli/v3"
	"golang.org/x/sys/unix"
)

type App struct {
	TransportConfig *string
	RoutingConfig   *routing.Config
}

func (app App) Run() error {
	trafficCopyWg := &sync.WaitGroup{}
	defer trafficCopyWg.Wait()

	tun, err := tundevice.New(app.RoutingConfig.TunDeviceName, app.RoutingConfig.TunDeviceIP)
	if err != nil {
		return fmt.Errorf("failed to create tun device: %w", err)
	}
	defer tun.Close()

	prevIPv6, err := ipv6.SetEnabled(false)
	if err != nil {
		return fmt.Errorf("failed to disable IPv6: %w", err)
	}
	defer ipv6.SetEnabled(prevIPv6)

	ss, err := outline_device.New(*app.TransportConfig)
	if err != nil {
		return fmt.Errorf("failed to create OutlineDevice: %w", err)
	}
	defer ss.Close()

	ss.Refresh()

	// Copy the traffic from tun device to OutlineDevice bidirectionally
	trafficCopyWg.Add(2)
	go func() {
		defer trafficCopyWg.Done()
		written, err := io.Copy(ss, tun)
		log.Printf("tun -> OutlineDevice stopped: %v %v\n", written, err)
	}()
	go func() {
		defer trafficCopyWg.Done()
		written, err := io.Copy(tun, ss)
		log.Printf("OutlineDevice -> tun stopped: %v %v\n", written, err)
	}()

	err = dns.SetSystemDNSServer(app.RoutingConfig.DNSServerIP)
	if err != nil {
		return fmt.Errorf("failed to configure system DNS: %w", err)
	}
	defer dns.RestoreSystemDNSServer()

	if err := routing.Start(ss.GetServerIP().String(), app.RoutingConfig); err != nil {
		return fmt.Errorf("failed to configure routing: %w", err)
	}
	defer routing.Stop(app.RoutingConfig.RoutingTableID)

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, unix.SIGTERM, unix.SIGHUP)
	s := <-sigc
	log.Printf("received %v, terminating...\n", s)
	return nil
}

func main() {
	appdata.Load()
	defer appdata.Save(appdata.AppData)

	cmd := &cli.Command{
		Name:  "tssc",
		Usage: "handle ss:// config urls",
		Commands: []*cli.Command{
			{
				Name:      "connect",
				Aliases:   []string{"c"},
				Usage:     "Establish connection",
				ArgsUsage: "<alias>",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					alias := cmd.Args().First()
					url := appdata.AppData.Urls[alias]

					app := &App{
						TransportConfig: &url,
						RoutingConfig: &routing.Config{
							TunDeviceName:        "outline233",
							TunDeviceIP:          "10.233.233.1",
							TunDeviceMTU:         1500, // todo: read this from netlink
							TunGatewayCIDR:       "10.233.233.2/32",
							RoutingTableID:       233,
							RoutingTablePriority: 23333,
							DNSServerIP:          "9.9.9.9",
						},
					}

					if err := app.Run(); err != nil {
						fmt.Println(err.Error())
						os.Exit(1)
					}

					return nil
				},
			},
			{
				Name:    "list",
				Aliases: []string{"l"},
				Usage:   "List all saved urls",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					for k, v := range appdata.AppData.Urls {
						fmt.Printf("%s :: %s\n", k, v)
					}
					return nil
				},
			},
			{
				Name:      "add",
				Aliases:   []string{"a"},
				Usage:     "Add url",
				ArgsUsage: "<alias> <ss://url>",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					if cmd.Args().Len() != 2 {
						fmt.Println("incorrect format")
						os.Exit(1)
					}

					alias := cmd.Args().Get(0)
					url := cmd.Args().Get(1)

					appdata.AppData.Urls[alias] = url

					return nil
				},
			},
		},
	}

	cmd.Run(context.Background(), os.Args)
}
