package main

import (
	"github.com/pion/stun"
)

func getExternalIP(stunServers []string) string {
	component := Component{
		Name:   "STUN Client",
		LogTag: "stun-client",
	}
	output := ""
	for _, stunServer := range stunServers {
		c, err := stun.Dial("udp4", stunServer)
		if err != nil {
			component.LogAlert("dial error [%s]", err)
			continue
		}
		if err = c.Do(stun.MustBuild(stun.TransactionID, stun.BindingRequest), func(res stun.Event) {
			if res.Error != nil {
				component.LogAlert("packet building error [%s]", res.Error)
				return
			}
			var xorAddr stun.XORMappedAddress
			if getErr := xorAddr.GetFrom(res.Message); getErr != nil {
				component.LogAlert("xorAddress error [%s]", getErr)
				return
			}
			output = xorAddr.IP.String()
		}); err != nil {
			component.LogAlert("error during STUN-do [%s]", err)
			continue
		}
		if err := c.Close(); err != nil {
			component.LogAlert("error closing STUN client [%s]", err)
		}

		if output != "" {
			break
		}
	}
	return output
}
