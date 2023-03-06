package cia_engine

import (
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
	//reqBodyBytes := new(bytes.Buffer)
	//json.NewEncoder(reqBodyBytes).Encode(cia)
	//jsonPretty, _ := tools.PrettyString(reqBodyBytes.Bytes())
	//fmt.Println(jsonPretty)
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
				var zoneInfos structure.ZoneInfos
				ok := false
				switch ciaCode.Commande {
				case "move":
					if cia.Api.TeamBlue {
						ok, err = cia.Algo.QuickMove(fmt.Sprintf("%d", zone.SecteurID), fmt.Sprintf("%d", zone.ZoneID))
					} else {
						ok, err = cia.Algo.Move(fmt.Sprintf("%d", zone.SecteurID), fmt.Sprintf("%d", zone.ZoneID))
					}
					break
				case "scan":
					_, res, errScan := cia.Algo.Scan()
					if errScan != nil {
						err = errScan
						break
					}
					err = json.Unmarshal(res, &zoneInfos)
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
					if err == nil {
						err = errors.New("erreur run commande")
					}
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
	tools.Title("Loop")
	if ciaCode.Good == "energy_seuil" && !cia.Status.Energy {
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
	for i := 0; i < count; i++ {
		currentCia := ciaCode.LoopCode[i]
		tools.Info(fmt.Sprintf(
			"\tcommande [%s] - instruction [%s] - action [%s]",
			currentCia.Commande,
			currentCia.Instruction,
			currentCia.Action,
		))
		var errCommande error
		var ok bool
		var res []byte
		var celluleData map[int]structure.CelluleData
		ActionTrapped := false
		switch currentCia.Commande {
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
		instructionSplit := strings.Split(currentCia.Instruction, "-")
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
		if currentCia.Action == "" {
			err = errors.New("need action")
			return
		}
		err = cia.LoopCodeAction(currentCia.Action, cellule, celluleData, ActionTrapped)
		if err != nil {
			return
		}
	}
	cia.Algo.Equilibrium()
	cia.Algo.ExplorationStop()
	return
}
func (cia *CiaEngine) CheckIsGood() (ok bool, err error) {
	tools.Title("Check is Good")
	ok = true
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
		ok = false
	}
	if energyTotal < seuilEnergy {
		cia.Status.Energy = true
		ok = false
	}
	tools.Info(fmt.Sprintf(
		"--- Report is good : valeur [%d] - seuil [%d] - [%t] | energy [%d] - seuil [%d] - [%t]",
		valeurTotal,
		seuilValeur,
		cia.Status.Rebuild,
		energyTotal,
		seuilEnergy,
		cia.Status.Energy,
	))
	return
}
func (cia *CiaEngine) LoopCodeAction(action string, cellule structure.CelluleInfos, celluleData map[int]structure.CelluleData, ActionTrapped bool) (err error) {
	tools.Title("Loop code action")
	cia.Next = false
	actionList := strings.Split(action, ",")
	nbrAction := len(actionList)
	for i := 0; i < nbrAction; i++ {
		actionSplit := strings.Split(strings.TrimSpace(actionList[i]), "-")
		tools.Info(fmt.Sprintf(
			"action-split = [%v] -  ActionTrapped [%t] - celluleTrapped [%t]",
			actionSplit,
			ActionTrapped,
			cellule.Trapped,
		))
		switch actionSplit[0] {
		case "capture":
			if !ActionTrapped && !cellule.Trapped {
				tools.Info("\t\t>> ICI <<")
				if actionSplit[1] == "energy" {
					count := len(celluleData)
					tools.Info(fmt.Sprintf("\t\t>> ICI << count [%d]", count))
					for j := 0; j < count; j++ {
						tools.Info(fmt.Sprintf("cellule id tentative capture [%d] index [$d]", cellule.ID, j))
						ok, _ := cia.Algo.CaptureCellEnergy(cellule.ID, j)
						if !ok {
							return
						}
					}
				} else if actionSplit[1] == "competence" {
					for _, data := range celluleData {
						if data.Competence {
							ok, _ := cia.Algo.CaptureCellData(cellule.ID, data.ID)
							if !ok {
								return
							}
						}
					}
				}
			}
			break
		case "destroy":
			if ActionTrapped && cellule.Trapped {
				cia.Algo.DestroyZone(cellule.ID, cellule.Valeur)
			}
			break
		case "next":
			break
		default:
			err = errors.New("code action not found")
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
						switch actionSplit[1] {
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
