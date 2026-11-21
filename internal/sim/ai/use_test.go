package ai

import (
	"encoding/json"
	"testing"

	"gtsandbox/internal/sim/cmds"
)

func TestOfferValidate(t *testing.T) {
	use := Step{Do: StepUse, Ticks: 10}
	item := &cmds.Spawn{Name: "cup"}
	for name, c := range map[string]struct {
		o  Offer
		ok bool
	}{
		"simple":           {Offer{Action: "a", Scale: "s", Duration: 5}, true},
		"sin duración":     {Offer{Action: "a", Scale: "s"}, false},
		"llevar y dejar":   {Offer{Action: "a", Scale: "s", Steps: []Step{{Do: StepTake, Item: item}, use, {Do: StepDrop}}}, true},
		"dejar sin llevar": {Offer{Action: "a", Scale: "s", Steps: []Step{use, {Do: StepDrop}}}, false},
		"dos use":          {Offer{Action: "a", Scale: "s", Steps: []Step{use, use}}, false},
		"paso desconocido": {Offer{Action: "a", Scale: "s", Steps: []Step{use, {Do: "fly"}}}, false},
		"item agente": {Offer{Action: "a", Scale: "s", Steps: []Step{use,
			{Do: StepSpawn, Item: &cmds.Spawn{Components: map[string]json.RawMessage{"agent": nil}}}}}, false},
	} {
		if err := c.o.Validate(); (err == nil) != c.ok {
			t.Errorf("%s: %v", name, err)
		}
	}
}
