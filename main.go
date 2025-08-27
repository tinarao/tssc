package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"sync"
	"tssc/internal/dns"
	"tssc/internal/ipv6"
	"tssc/internal/outline_device"
	"tssc/internal/routing"
	tundevice "tssc/internal/tun_device"

	"golang.org/x/sys/unix"
)

type App struct {
	TransportConfig *string
	RoutingConfig   *routing.Config
}

func (app App) Run() error {
	// this WaitGroup must Wait() after tun is closed
	trafficCopyWg := &sync.WaitGroup{}
	defer trafficCopyWg.Wait()

	tun, err := tundevice.New(app.RoutingConfig.TunDeviceName, app.RoutingConfig.TunDeviceIP)
	if err != nil {
		return fmt.Errorf("failed to create tun device: %w", err)
	}
	defer tun.Close()

	// disable IPv6 before resolving Shadowsocks server IP
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

	app := App{
		TransportConfig: flag.String("transport", "", "Transport config"),
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
	flag.Parse()

	if err := app.Run(); err != nil {
		log.Printf("%v\n", err)
	}

	// u := "penis"
	//
	// device, err := outline_device.New(u)
	// if err != nil {
	// 	panic(err.Error())
	// }
	//
	// defer device.Close()
	// ss.Refresh()

	// conf := ResolveShadowsocksServerIPFromConfig
}
