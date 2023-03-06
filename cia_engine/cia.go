package cia_engine

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/greg198584/client-ag3/algo"
	"github.com/greg198584/client-ag3/tools"
	"strings"
)

type CiaEngine struct {
	ScriptName    string  `json:"script_name"`
	ProgrammeName string  `json:"programme_name"`
	AutoConnect   bool    `json:"auto_connect"`
	Api           Api     `json:"api"`
	LoopCIA       LoopCia `json:"loop_cia"`
}

type Api struct {
	Url      string `json:"url"`
	TeamBlue bool   `json:"team_blue"`
	TeamRed  bool   `json:"team_red"`
}

type LoopCia struct {
	LoopParams LoopParams `json:"loop_params"`
	LoopCode   []CIA      `json:"loop_code"`
}

type CIA struct {
	Commande    string `json:"commande"`
	Instruction string `json:"instruction"`
	Action      string `json:"action"`
	LoopCode    []SCIA `json:"loop_code"`
}

type SCIA struct {
	Commande    string `json:"commande"`
	Instruction string `json:"instruction"`
	Action      string `json:"action"`
}

type LoopParams struct {
	Stop   bool `json:"stop"`
	Random bool `json:"random"`
}

func _LoadScript(name string) (cia *CiaEngine, err error) {
	file, err := tools.GetJsonFile(fmt.Sprintf("./script/%s.json", name))
	if err != nil {
		return cia, err
	}
	err = json.Unmarshal(file, &cia)
	if err != nil {
		return cia, err
	}
	return
}
func New(name string) (cia *CiaEngine, err error) {
	cia, err = _LoadScript(name)
	if err != nil {
		return
	}
	cia.ScriptName = name
	return
}

func (cia *CiaEngine) Run() (err error) {
	tools.Info(fmt.Sprintf("Run script [%s] - programme [%s]", cia.ScriptName, cia.ProgrammeName))
	var current *algo.Algo
	if cia.Api.TeamBlue {
		current, err = algo.NewAlgoBlueTeam(cia.ProgrammeName, cia.Api.Url)
	} else {
		current, err = algo.NewAlgo(cia.ProgrammeName, cia.Api.Url)
	}
	if err != nil {
		return
	}
	err = current.GetStatusGrid()
	if err != nil {
		return
	}
	tools.PrintInfosGrille(current.InfosGrid)
	reqBodyBytes := new(bytes.Buffer)
	json.NewEncoder(reqBodyBytes).Encode(cia)
	jsonPretty, _ := tools.PrettyString(reqBodyBytes.Bytes())
	fmt.Println(jsonPretty)
	if cia.AutoConnect {
		if ok, errLoad := current.LoadProgramme(); !ok {
			err = errLoad
			return
		}
	} else {
		if ok, errInfo := current.GetInfosProgramme(); !ok {
			err = errInfo
			return
		}
	}
	for _, zone := range current.InfosGrid.Zones {
		if zone.Status {
			tools.Title(fmt.Sprintf("Zone [%d][%d]", zone.SecteurID, zone.ZoneID))
			count := len(cia.LoopCIA.LoopCode)
			for i := 0; i < count; i++ {
				ciaCode := cia.LoopCIA.LoopCode[i]
				tools.Info(fmt.Sprintf(
					"\tcommande [%s] - instruction [%s] - action [%s]",
					ciaCode.Commande,
					ciaCode.Instruction,
					ciaCode.Action,
				))
				var errCommande error
				var ok bool
				switch ciaCode.Commande {
				case "move":
					if cia.Api.TeamBlue {
						ok, errCommande = current.QuickMove(fmt.Sprintf("%d", zone.SecteurID), fmt.Sprintf("%d", zone.ZoneID))
					} else {
						ok, errCommande = current.Move(fmt.Sprintf("%d", zone.SecteurID), fmt.Sprintf("%d", zone.ZoneID))
					}
				}
				if !ok {
					err = errCommande
					return
				}
				instructionSplit := strings.Split(ciaCode.Instruction, "-")
				tools.Info(fmt.Sprintf("instruction-split = [%v]", instructionSplit))

				if ciaCode.Action == "" {
					err = errors.New("action not found")
					return
				}
			}
		}
	}
	return
}
