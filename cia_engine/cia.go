package cia_engine

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/greg198584/client-ag3/algo"
	"github.com/greg198584/client-ag3/structure"
	"github.com/greg198584/client-ag3/tools"
	"net/http"
	"strings"
	"time"
)

type CiaEngine struct {
	Algo           *algo.Algo
	ScriptName     string        `json:"script_name"`
	ProgrammeName  string        `json:"programme_name"`
	AutoConnect    bool          `json:"auto_connect"`
	ZoneActif      bool          `json:"zone_actif"`
	ErrorSeuil     int           `json:"error_seuil"`
	Api            Api           `json:"api"`
	LoopCIA        LoopCia       `json:"loop_cia"`
	Next           bool          `json:"next"`
	Status         Status        `json:"status"`
	Mem            GrilleZoneMem `json:"mem"`
	ConnexionError int           `json:"connexion_error"`
}

type GrilleZoneMem struct {
	ZoneInfos        structure.ZoneInfos
	Targets          []string
	Cellules         map[int]*structure.Cellule
	ShellCodeTarget  []structure.ShellcodeData
	MaxValeurCellule int
	MaxEnergyCellule int
}

type Status struct {
	Energy    bool `json:"energy"`
	Rebuild   bool `json:"rebuild"`
	Attack    bool `json:"attack"`
	ShellCode bool `json:"shellcode"`
	FlagFound bool `json:"flag_found"`
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
type CelluleDataList struct {
	CelluleID   int
	Valeur      int
	Trapped     bool
	CelluleData map[int]structure.CelluleData `json:"cellule_data"`
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
	time.Sleep(algo.TIME_MILLISECONDE * time.Millisecond)
	cia.Mem.Targets = []string{}
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
		if ok, errLoad := cia.Algo.LoadProgramme(cia.Api.TeamBlue); !ok {
			err = errLoad
			return
		}
	}
	if ok, errInfo := cia.Algo.GetInfosProgramme(); !ok {
		err = errInfo
		return
	}
	cia.Mem.MaxValeurCellule = (cia.Algo.Psi.Programme.Level * algo.MAX_VALEUR) * algo.MAX_CELLULES
	cia.Mem.MaxEnergyCellule = ((cia.Algo.Psi.Programme.Level * algo.MAX_VALEUR) * algo.MAX_CELLULES) * 2
	cia.Algo.ExplorationStop()
	for _, zone := range cia.Algo.InfosGrid.Zones {
		if cia.Algo.StatusCode == http.StatusUnauthorized {
			cia.ConnexionError++
			tools.Fail(fmt.Sprintf("add http error 401 - nbr error 401 [%d]", cia.ConnexionError))
		}
		if cia.ConnexionError == cia.ErrorSeuil {
			if cia.AutoConnect {
				if ok, errLoad := cia.Algo.LoadProgramme(cia.Api.TeamBlue); !ok {
					err = errLoad
					return
				}
			}
			if ok, errInfo := cia.Algo.GetInfosProgramme(); !ok {
				err = errInfo
				return
			}
			cia.ConnexionError = 0
		}
		if zone.Status {
			if zone.Actif == false && cia.ZoneActif {
				continue
			}
			count := len(cia.LoopCIA.LoopCode)
			tools.Title(fmt.Sprintf("Zone [%d][%d] - cia [%d]", zone.SecteurID, zone.ZoneID, count))
			cia.Next = true
			time.Sleep(algo.TIME_MILLISECONDE * time.Millisecond)
			for i := 0; i < count; i++ {
				tools.Info(fmt.Sprintf(
					"cia [%s] [%s] [%s]",
					cia.LoopCIA.LoopCode[i].Commande,
					cia.LoopCIA.LoopCode[i].Instruction,
					cia.LoopCIA.LoopCode[i].Action,
				))
				forceNext := false
				if cia.Next == false {
					tools.Warning("next false stop current cia")
					break
				}
				cia.Next = false
				ciaCode := cia.LoopCIA.LoopCode[i]
				tools.Info(fmt.Sprintf(
					"\tcommande [%s] - instruction [%s] - action [%s]",
					ciaCode.Commande,
					ciaCode.Instruction,
					ciaCode.Action,
				))
				ok := false
				switch ciaCode.Commande {
				case "move":
					if cia.Algo.Psi.Navigation || cia.Algo.Psi.Programme.Exploration {
						ok = true
						break
					}
					ok, err = cia.Algo.QuickMove(fmt.Sprintf("%d", zone.SecteurID), fmt.Sprintf("%d", zone.ZoneID))
					//if cia.Api.TeamBlue {
					//	ok, err = cia.Algo.QuickMove(fmt.Sprintf("%d", zone.SecteurID), fmt.Sprintf("%d", zone.ZoneID))
					//} else {
					//	ok, err = cia.Algo.Move(fmt.Sprintf("%d", zone.SecteurID), fmt.Sprintf("%d", zone.ZoneID))
					//}
					break
				case "scan":
					_, res, errScan := cia.Algo.Scan()
					if errScan != nil {
						err = errScan
						break
					}
					err = json.Unmarshal(res, &cia.Mem.ZoneInfos)
					if err != nil {
						break
					}
					ok = true
					break
				case "move_zt":
					if cia.Status.FlagFound {
						ok, err = cia.Algo.QuickMove(
							fmt.Sprintf("%d", cia.Algo.InfosGrid.ZoneTransfert.SecteurID),
							fmt.Sprintf("%d", cia.Algo.InfosGrid.ZoneTransfert.ZoneID),
						)
						//if cia.Api.TeamBlue {
						//	ok, err = cia.Algo.QuickMove(
						//		fmt.Sprintf("%d", cia.Algo.InfosGrid.ZoneTransfert.SecteurID),
						//		fmt.Sprintf("%d", cia.Algo.InfosGrid.ZoneTransfert.ZoneID),
						//	)
						//} else {
						//	ok, err = cia.Algo.Move(
						//		fmt.Sprintf("%d", cia.Algo.InfosGrid.ZoneTransfert.SecteurID),
						//		fmt.Sprintf("%d", cia.Algo.InfosGrid.ZoneTransfert.ZoneID),
						//	)
						//}
					} else {
						forceNext = true
					}
					ok = true
				case "rebuild":
					if cia.Status.Rebuild && cia.LoopCIA.LoopParams.Rebuild {
						cia.Algo.GetInfosProgramme()
						ok = true
					} else {
						ok = true
						forceNext = true
					}
					break
				case "attack":
					if cia.LoopCIA.LoopParams.Attack {
						cia.Status.Attack = true
						ok = true
					} else {
						ok = true
						forceNext = true
					}
					break
				case "shellcode":
					if cia.LoopCIA.LoopParams.ShellCode {
						cia.Status.ShellCode = true
						ok = true
					} else {
						ok = true
						forceNext = true
					}
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
					tools.Title("wait")
					err = cia.Wait(instructionSplit[1], instructionSplit[2])
					break
				case "loop":
					tools.Title("loop")
					err = cia.Loop(ciaCode.Code, instructionSplit[1], instructionSplit[2])
					break
				case "nop":
					tools.Title("nop")
					break
				case "check":
					tools.Title("check")
					if cia.Status.Rebuild {
						cia.Mem.Cellules = make(map[int]*structure.Cellule)
						tools.Warning("> rebuild")
						for _, cellule := range cia.Algo.Psi.Programme.Cellules {
							if cellule.Status == false {
								cia.Mem.Cellules[cellule.ID] = cellule
								cia.Mem.Cellules[cellule.ID].Energy = (cia.Mem.MaxValeurCellule / 4) - cellule.Valeur + 1
							}
						}
					}
					if cia.Status.Attack {
						tools.Warning("> attack")
						for celluleID := 0; celluleID < algo.MAX_CELLULES; celluleID++ {
							cia.Algo.Explore(celluleID)
							logs, _ := cia.Algo.GetLog(celluleID)
							for _, log := range logs {
								cia.Mem.Targets = append(cia.Mem.Targets, log.PID)
							}

						}
					}
					break
				case "active":
					if cia.Status.ShellCode {
						_, cia.Mem.ShellCodeTarget, err = cia.Algo.ShellCode()
					}
					break
				default:
					err = errors.New("instruction not found")
					break
				}
				if err != nil {
					return
				}
				if ciaCode.Action == "" {
					err = errors.New("need action")
					return
				}
				if forceNext {
					cia.Next = true
					tools.Warning("force next")
				} else {
					err = cia.Action(ciaCode)
					if err != nil {
						return
					}
					tools.Success(fmt.Sprintf("End action %s - %s", ciaCode.Commande, ciaCode.Action))
				}
			}
		}
	}
	return
}
func (cia *CiaEngine) Loop(ciaCode CiaCode, value string, condition string) (err error) {
	tools.Title("Loop")
	if ciaCode.Good == "energy_seuil" && !cia.Status.Energy {
		return
	}
	switch value {
	case "cellule":
		for _, cellule := range cia.Mem.ZoneInfos.Cellules {
			switch condition {
			case "code":
				err = cia.RunCodeCellule(ciaCode, cellule)
				break
			default:
				err = errors.New("erreur condition instruction")
				break
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
	tools.Info(fmt.Sprintf("Run code cellule [%s] - celluleID [%d]", ciaCode.Name, cellule.ID))
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
	cia.Algo.CleanLogAll()
	cia.Algo.Equilibrium()
	cia.Algo.ExplorationStop()
	return
}
func (cia *CiaEngine) RunCode(ciaCode CiaCode) (err error) {
	tools.Info(fmt.Sprintf("Run code [%s]", ciaCode.Name))
	count := len(ciaCode.LoopCode)
	//var celluleDataList map[int]CelluleDataList
	celluleDataList := make(map[int]CelluleDataList)
	var zoneInfos structure.ZoneInfos
	for i := 0; i < count; i++ {
		currentCia := ciaCode.LoopCode[i]
		tools.Info(fmt.Sprintf(
			"\tcommande [%s] - instruction [%s] - action [%s]",
			currentCia.Commande,
			currentCia.Instruction,
			currentCia.Action,
		))
		var errCommande error
		ok := false
		ActionTrapped := false
		ActionCaptureFlag := false
		switch currentCia.Commande {
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
		case "explore":
			for _, cellule := range zoneInfos.Cellules {
				_, res, errExplore := cia.Algo.Explore(cellule.ID)
				if errExplore != nil {
					err = errExplore
					break
				}
				var celluleData map[int]structure.CelluleData
				err = json.Unmarshal(res, &celluleData)
				if err != nil {
					break
				}
				celluleDataList[cellule.ID] = CelluleDataList{
					CelluleID:   cellule.ID,
					Valeur:      cellule.Valeur,
					Trapped:     cellule.Trapped,
					CelluleData: celluleData,
				}
				ok = true
				break
			}
		case "searchflag":
			ok = true
			break
		case "pushflag":
			ok, err = cia.Algo.PushFlag()
			return
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
			if instructionSplit[1] == "flag" && instructionSplit[2] == "true" {
				ActionCaptureFlag = true
			}
			break
		case "loop":
			if instructionSplit[1] == "cellule" && instructionSplit[2] == "next" {
				break
			} else {
				err = errors.New("need 3 instructions")
				return
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
		cia.ActionCode(currentCia.Action, celluleDataList, ActionTrapped, ActionCaptureFlag)
	}
	cia.Algo.GetInfosProgramme()
	cia.Algo.CleanLogAll()
	cia.Algo.Equilibrium()
	cia.Algo.ExplorationStop()
	return
}
func (cia *CiaEngine) ActionCode(action string, celluleDataList map[int]CelluleDataList, ActionTrapped bool, ActionCaptureFlag bool) {
	actionList := strings.Split(action, ",")
	nbrAction := len(actionList)
	for i := 0; i < nbrAction; i++ {
		actionSplit := strings.Split(strings.TrimSpace(actionList[i]), "-")
		switch actionSplit[0] {
		case "next":
			return
		case "destroy":
			for _, cellule := range celluleDataList {
				if ActionTrapped && cellule.Trapped {
					cia.Algo.DestroyZone(cellule.CelluleID, cellule.Valeur)
				}
			}
		case "capture":
			if ActionCaptureFlag {
				for _, cellule := range celluleDataList {
					for _, data := range cellule.CelluleData {
						tools.Info(fmt.Sprintf("cellule ID: [%d] - data [%d]", cellule.CelluleID, data.ID))
						if data.IsFlag {
							tools.Success("flag found")
							cia.Algo.CaptureCellData(cellule.CelluleID, data.ID)
							break
						}
					}
				}
				tools.Fail("flag not found")
			}
		default:
			return
		}
	}
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
	if valeurTotal < cia.Mem.MaxValeurCellule {
		cia.Status.Rebuild = true
		ok = false
	}
	if energyTotal < cia.Mem.MaxEnergyCellule {
		cia.Status.Energy = true
		ok = false
	}
	tools.Info(fmt.Sprintf(
		"--- Report is good [%t]: valeur [%d] - seuil [%d] - [%t] | energy [%d] - seuil [%d] - [%t]",
		ok,
		valeurTotal,
		cia.Mem.MaxValeurCellule,
		cia.Status.Rebuild,
		energyTotal,
		cia.Mem.MaxEnergyCellule,
		cia.Status.Energy,
	))
	cia.Status.FlagFound = false
	for _, cellule := range cia.Algo.Psi.Programme.Cellules {
		for _, data := range cellule.Datas {
			if data.IsFlag {
				cia.Status.FlagFound = true
			}
		}
	}
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
				if actionSplit[1] == "energy" {
					for _, data := range celluleData {
						tools.Info(fmt.Sprintf("cellule id tentative capture [%d] index [%d]", cellule.ID, data.ID))
						if data.Energy > 0 {
							ok, _ := cia.Algo.CaptureCellEnergy(cellule.ID, data.ID)
							if !ok {
								tools.Warning("stop capture energy")
								break
							}
						}
					}
				} else if actionSplit[1] == "competence" {
					for _, data := range celluleData {
						tools.Info(fmt.Sprintf("cellule id tentative capture [%d] index [%d]", cellule.ID, data.ID))
						if data.Competence {
							ok, _ := cia.Algo.CaptureCellData(cellule.ID, data.ID)
							if !ok {
								tools.Warning("stop capture competence")
								break
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
func (cia *CiaEngine) Action(ciaCode CIA) (err error) {
	tools.Info(fmt.Sprintf("Run action [%s]", ciaCode.Action))
	actionList := strings.Split(ciaCode.Action, ",")
	nbrAction := len(actionList)
	for i := 0; i < nbrAction; i++ {
		cia.Next = false
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
					ok, errCheck := cia.CheckIsGood()
					err = errCheck
					switch actionSplit[1] {
					case "is":
						if ok {
							cia.Next = true
						} else {
							cia.Next = false
						}
						break
					case "isnot":
						if ok {
							cia.Next = false
						} else {
							cia.Next = true
						}
						break
					default:
						break
					}
					break
				default:
					err = errors.New("action fail condition")
					break
				}
			}
			return
		case "code":
			err = cia.RunCode(ciaCode.Code)
			break
		case "rebuild":
			tools.Warning("\n\t ---> REBUILD ...\n")
			for _, cellule := range cia.Mem.Cellules {
				tools.Info(fmt.Sprintf("\n\t\tcellule id [%d] [%d] [%d]\n", cellule.ID, cellule.Valeur, cellule.Energy))
				ok, _, _ := cia.Algo.Rebuild(cellule.ID, cia.Algo.Psi.Programme.ID, cellule.Energy)
				if ok {
					delete(cia.Mem.Cellules, cellule.ID)
				}
			}
			break
		case "attack":
			tools.Warning("\n\t ---> ATTACK ...\n")
			condition := actionSplit[1]
			switch condition {
			case "max":
				for _, targetID := range cia.Mem.Targets {
					for _, cellule := range cia.Algo.Psi.Programme.Cellules {
						if cellule.Status {
							cia.Algo.Destroy(cellule.ID, targetID, cellule.Valeur)
						}
					}
				}
				break
			default:
				break
			}
		case "active":
			tools.Warning("\n\t ---> ACTIVE SHELLCODE ...\n")
			if cia.Status.ShellCode {
				tools.PrintShellCodeData(cia.Mem.ShellCodeTarget)
				for _, shellCodeTarget := range cia.Mem.ShellCodeTarget {
					if shellCodeTarget.BlueTeam && cia.Api.TeamBlue {
						tools.Warning("target blue team")
					} else {
						tools.Fail(fmt.Sprintf("active shellcode target [%s]", shellCodeTarget.PID))
						cia.Algo.ActiveShellCode(shellCodeTarget.PID, shellCodeTarget.Shellcode)
					}
				}
			}
			break
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
