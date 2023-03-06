package cia_engine

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/greg198584/client-ag3/algo"
	"github.com/greg198584/client-ag3/structure"
	"github.com/greg198584/client-ag3/tools"
	"strings"
	"time"
)

type CiaEngine struct {
	Algo          *algo.Algo
	ScriptName    string  `json:"script_name"`
	ProgrammeName string  `json:"programme_name"`
	AutoConnect   bool    `json:"auto_connect"`
	Api           Api     `json:"api"`
	LoopCIA       LoopCia `json:"loop_cia"`
	Next          bool    `json:"next"`
	Status        Status  `json:"status"`
}

type Status struct {
	Energy    bool `json:"energy"`
	Rebuild   bool `json:"rebuild"`
	Attack    bool `json:"attack"`
	ShellCode bool `json:"shellcode"`
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
	Commande    string  `json:"commande"`
	Instruction string  `json:"instruction"`
	Action      string  `json:"action"`
	Code        CiaCode `json:"code"`
}

type CiaCode struct {
	Name     string `json:"name"`
	Good     string `json:"good"`
	LoopCode []CIA  `json:"loop_code"`
}

type LoopParams struct {
	Stop        bool `json:"stop"`
	Random      bool `json:"random"`
	EnergySeuil int  `json:"energy_seuil"`
	Rebuild     bool `json:"rebuild"`
	Attack      bool `json:"attack"`
	ShellCode   bool `json:"shellcode"`
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
func NewFromSpeedCode(speedCode string) (cia *CiaEngine, err error) {
	//cia, err = _LoadScript(name)
	//if err != nil {
	//	return
	//}
	//cia.ScriptName = name
	return
}
func (cia *CiaEngine) Run() (err error) {
	tools.Info(fmt.Sprintf("Run script [%s] - programme [%s]", cia.ScriptName, cia.ProgrammeName))
	if cia.Api.TeamBlue {
		cia.Algo, err = algo.NewAlgoBlueTeam(cia.ProgrammeName, cia.Api.Url)
	} else {
		cia.Algo, err = algo.NewAlgo(cia.ProgrammeName, cia.Api.Url)
	}
	if err != nil {
		return
	}
	err = cia.Algo.GetStatusGrid()
	if err != nil {
		return
	}
	tools.PrintInfosGrille(cia.Algo.InfosGrid)
	reqBodyBytes := new(bytes.Buffer)
	json.NewEncoder(reqBodyBytes).Encode(cia)
	jsonPretty, _ := tools.PrettyString(reqBodyBytes.Bytes())
	fmt.Println(jsonPretty)
	if cia.AutoConnect {
		if ok, errLoad := cia.Algo.LoadProgramme(); !ok {
			err = errLoad
			return
		}
	} else {
		if ok, errInfo := cia.Algo.GetInfosProgramme(); !ok {
			err = errInfo
			return
		}
	}
	for _, zone := range cia.Algo.InfosGrid.Zones {
		if zone.Status {
			tools.Title(fmt.Sprintf("Zone [%d][%d]", zone.SecteurID, zone.ZoneID))
			count := len(cia.LoopCIA.LoopCode)
			cia.Next = true
			for i := 0; i < count; i++ {
				if cia.Next == false {
					break
				}
				ciaCode := cia.LoopCIA.LoopCode[i]
				tools.Info(fmt.Sprintf(
					"\tcommande [%s] - instruction [%s] - action [%s]",
					ciaCode.Commande,
					ciaCode.Instruction,
					ciaCode.Action,
				))
				var errCommande error
				var ok bool
				var zoneInfos structure.ZoneInfos
				switch ciaCode.Commande {
				case "move":
					if cia.Api.TeamBlue {
						ok, errCommande = cia.Algo.QuickMove(fmt.Sprintf("%d", zone.SecteurID), fmt.Sprintf("%d", zone.ZoneID))
					} else {
						ok, errCommande = cia.Algo.Move(fmt.Sprintf("%d", zone.SecteurID), fmt.Sprintf("%d", zone.ZoneID))
					}
					break
				case "scan":
					ok, res, errScan := cia.Algo.Scan()
					if !ok {
						errCommande = errScan
						break
					}
					errCommande = json.Unmarshal(res, &zoneInfos)
					if err != nil {
						break
					}
					ok = true
					break
				default:
					err = errors.New("commande not found")
					return
				}
				if !ok {
					err = errCommande
					return
				}
				instructionSplit := strings.Split(ciaCode.Instruction, "-")
				tools.Info(fmt.Sprintf("instruction-split = [%v]", instructionSplit))
				if len(instructionSplit) != 3 {
					err = errors.New("need 3 instructions")
					return
				}
				switch instructionSplit[0] {
				case "wait":
					err = cia.Wait(instructionSplit[1], instructionSplit[2])
					break
				case "loop":
					err = cia.Loop(ciaCode.Code, instructionSplit[1], instructionSplit[2], zoneInfos)
					break
				default:
					err = errors.New("action not found")
					break
				}
				if err != nil {
					return
				}
				if ciaCode.Action == "" {
					err = errors.New("need action")
					return
				}
				err = cia.Action(ciaCode.Action)
				if err != nil {
					return
				}
			}
		}
	}
	return
}
func (cia *CiaEngine) Loop(ciaCode CiaCode, value string, condition string, zoneInfos structure.ZoneInfos) (err error) {
	if ciaCode.Good == "energy_seuil" && cia.Status.Energy == false {
		return
	}
	switch value {
	case "cellule":
		for _, cellule := range zoneInfos.Cellules {
			if condition == "code" {
				err = cia.RunCodeCellule(ciaCode, cellule)
			}
			if err != nil {
				break
			}

		}
		break
	default:
		err = errors.New("erreur value instruction")
		break
	}
	return
}
func (cia *CiaEngine) RunCodeCellule(ciaCode CiaCode, cellule structure.CelluleInfos) (err error) {
	tools.Info(fmt.Sprintf("Run code [%s] - celluleID [%d]", ciaCode.Name, cellule.ID))
	count := len(ciaCode.LoopCode)
	cia.Next = true
	for i := 0; i < count; i++ {
		if cia.Next == false {
			break
		}
		ciaCode := cia.LoopCIA.LoopCode[i]
		tools.Info(fmt.Sprintf(
			"\tcommande [%s] - instruction [%s] - action [%s]",
			ciaCode.Commande,
			ciaCode.Instruction,
			ciaCode.Action,
		))
		var errCommande error
		var ok bool
		var res []byte
		var celluleData structure.Cellule
		ActionTrapped := false
		switch ciaCode.Commande {
		case "explore":
			ok, res, err = cia.Algo.Explore(cellule.ID)
			if ok {
				errCommande = json.Unmarshal(res, &celluleData)
				if err != nil {
					break
				}
			}
			break
		}
		if !ok {
			err = errCommande
			return
		}
		instructionSplit := strings.Split(ciaCode.Instruction, "-")
		tools.Info(fmt.Sprintf("instruction-split = [%v]", instructionSplit))
		if len(instructionSplit) != 3 {
			err = errors.New("need 3 instructions")
			return
		}
		switch instructionSplit[0] {
		case "check":
			if instructionSplit[1] == "trapped" && instructionSplit[2] == "true" {
				ActionTrapped = true
			}
			break
		}
		if err != nil {
			return
		}
		if ciaCode.Action == "" {
			err = errors.New("need action")
			return
		}
		err = cia.LoopCodeAction(ciaCode.Action, celluleData, ActionTrapped)
		if err != nil {
			return
		}
	}
	return
}
func (cia *CiaEngine) CheckIsGood() (ok bool, err error) {
	if ok, err = cia.Algo.GetInfosProgramme(); !ok {
		return
	}
	energyTotal := 0
	valeurTotal := 0
	for _, cellule := range cia.Algo.Psi.Programme.Cellules {
		energyTotal += cellule.Energy
		valeurTotal += cellule.Valeur
	}
	seuilValeur := (cia.Algo.Psi.Programme.Level * algo.MAX_VALEUR) * algo.MAX_CELLULES
	seuilEnergy := ((cia.Algo.Psi.Programme.Level * algo.MAX_VALEUR) * algo.MAX_CELLULES) * 10
	if valeurTotal < seuilValeur {
		cia.Status.Rebuild = true
	}
	if energyTotal < seuilEnergy {
		cia.Status.Energy = true
	}
	return
}
func (cia *CiaEngine) LoopCodeAction(action string, celluleData structure.Cellule, ActionTrapped bool) (err error) {
	cia.Next = false
	actionList := strings.Split(action, ",")
	nbrAction := len(actionList)
	for i := 0; i < nbrAction; i++ {
		actionSplit := strings.Split(strings.TrimSpace(actionList[i]), "-")
		switch actionSplit[0] {
		case "capture":
			if ActionTrapped && celluleData.Trapped {
				count := len(celluleData.Datas)
				for i := 0; i < count; i++ {
					ok, _ := cia.Algo.CaptureCellEnergy(celluleData.ID, i)
					if !ok {
						return
					}
				}
			}
			break
		case "destroy":
			if ActionTrapped && celluleData.Trapped {
				cia.Algo.DestroyZone(celluleData.ID, celluleData.Valeur)
			}
			break
		default:
			err = errors.New("action not found")
			return
		}
	}
	return
}
func (cia *CiaEngine) Action(action string) (err error) {
	cia.Next = false
	actionList := strings.Split(action, ",")
	nbrAction := len(actionList)
	for i := 0; i < nbrAction; i++ {
		actionSplit := strings.Split(strings.TrimSpace(actionList[i]), "-")
		switch actionSplit[0] {
		case "next":
			cia.Next = true
			return
		case "stop":
			if len(actionSplit) > 2 {
				condition := actionSplit[2]
				switch condition {
				case "good":
					if ok, errCheck := cia.CheckIsGood(); !ok {
						err = errCheck
						switch actionList[1] {
						case "is":
							cia.Next = true
							break
						case "isnot":
							break
						default:
							break
						}
					}
					break
				default:
					err = errors.New("action fail condition")
					break
				}
			}
			return
		default:
			err = errors.New("action not found")
			return
		}
	}
	return
}
func (cia *CiaEngine) Wait(value string, condition string) (err error) {
	for {
		time.Sleep(algo.TIME_MILLISECONDE * time.Millisecond)
		tools.Success(fmt.Sprintf("+++ waiting [%s] - [%s]", value, condition))
		if ok, errInfos := cia.Algo.GetInfosProgramme(); !ok {
			err = errInfos
			return
		}
		switch value {
		case "navigation":
			if cia.Algo.Psi.Navigation == false && condition == "false" {
				return
			}
			if cia.Algo.Psi.Navigation && condition == "true" {
				return
			}
		}
	}
}
