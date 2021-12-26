package main

import (
	"github.com/pion/stun"
)

func getExternalIP(stunServers []string) string {
	logger := Logger{
		Name: "stun-client",
		ID:   STUNClientID,
	}
	output := ""
	for _, stunServer := range stunServers {
		c, err := stun.Dial("udp4", stunServer)
		if err != nil {
			logger.LogAlert("dial error [%s]", err)
			continue
		}
		if err = c.Do(stun.MustBuild(stun.TransactionID, stun.BindingRequest), func(res stun.Event) {
			if res.Error != nil {
				logger.LogAlert("packet building error [%s]", res.Error)
				return
			}
			var xorAddr stun.XORMappedAddress
			if getErr := xorAddr.GetFrom(res.Message); getErr != nil {
				logger.LogAlert("xorAddress error [%s]", getErr)
				return
			}
			output = xorAddr.IP.String()
		}); err != nil {
			logger.LogAlert("error during STUN-do [%s]", err)
			continue
		}
		if err := c.Close(); err != nil {
			logger.LogAlert("error closing STUN client [%s]", err)
		}

		if output != "" {
			break
		}
	}
	return output
}
