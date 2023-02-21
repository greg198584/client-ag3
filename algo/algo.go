package algo

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/greg198584/client-ag3/api"
	"github.com/greg198584/client-ag3/structure"
	"github.com/greg198584/client-ag3/tools"
	"github.com/logrusorgru/aurora"
	"net/http"
	"os"
)

const (
	TIME_MILLISECONDE = 5000
	ENERGY_MAX_ATTACK = 10
	MAX_CELLULES      = 9
	MAX_VALEUR        = 100
)

type Algo struct {
	Name       string
	ApiUrl     string
	ID         string
	Pc         structure.ProgrammeContainer
	InfosGrid  structure.GridInfos
	Psi        structure.ProgrammeStatusInfos
	StatusCode int
}

func _LoadProgramme(name string) (pc structure.ProgrammeContainer, err error) {
	pc, err = _GetProgrammeFile(name)
	return
}
func _LoadProgrammeBlueTeam(name string, apiteam string) (psi structure.ProgrammeStatusInfos, pc structure.ProgrammeContainer, err error) {
	pc, err = _GetProgrammeFile(name)
	if pc.ID == "" || err != nil {
		//tools.Fail(fmt.Sprintf("no content [%s][%v]", name, pc))
		pc, _ = _CreateProgramme(name, apiteam)
		return
	} else {
		url := ""
		if apiteam == "a" {
			url = api.API_URL_A
		} else {
			url = api.API_URL_B
		}
		reqBodyBytes := new(bytes.Buffer)
		json.NewEncoder(reqBodyBytes).Encode(pc)
		res, statusCode, _ := api.RequestApi(
			"POST",
			fmt.Sprintf("%s/%s", url, api.ROUTE_LOAD_PROGRAMME_BLUE_TEAM),
			reqBodyBytes.Bytes(),
		)
		if statusCode == http.StatusCreated || statusCode == http.StatusOK {
			_ = json.Unmarshal(res, &psi)
		} else {
			pc, _ = _CreateProgramme(name, apiteam)
		}
	}
	return
}
func _CreateProgramme(name string, apiteam string) (programme structure.ProgrammeContainer, err error) {
	tools.Title(fmt.Sprintf("création programme [%s]", name))
	if _IsExistFile(name) == false {
		url := ""
		if apiteam == "a" {
			url = api.API_URL_A
		} else {
			url = api.API_URL_B
		}
		res, statusCode, err := api.RequestApi(
			"GET",
			fmt.Sprintf("%s/%s/%s", url, api.ROUTE_NEW_PROGRAMME, name),
			nil,
		)
		if err != nil {
			return programme, err
		}
		if statusCode == http.StatusCreated {
			err = json.Unmarshal(res, &programme)
			tools.CreateJsonFile(fmt.Sprintf("%s.json", name), programme)
			tools.Success("backup OK")
		} else {
			err = errors.New("erreur creation programme")
			return programme, err
		}
	} else {
		tools.Warning(fmt.Sprintf("programme file exist"))
	}
	return programme, err
}
func _IsExistFile(name string) bool {
	filename := fmt.Sprintf("%s.json", name)
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return false
	} else {
		return true
	}
}
func _GetProgrammeFile(name string) (pc structure.ProgrammeContainer, err error) {
	file, err := tools.GetJsonFile(fmt.Sprintf("%s.json", name))
	if err != nil {
		return pc, err
	}
	err = json.Unmarshal(file, &pc)
	if err != nil {
		return pc, err
	}
	return
}
func NewAlgo(name string, apiteam string) (algo *Algo, err error) {
	tools.Title(fmt.Sprintf("chargement programme [%s]", name))
	pc, err := _LoadProgramme(name)
	if err != nil {
		return
	} else {
		tools.Info(fmt.Sprintf("pc id = [%s]", pc.ID))
		algo = &Algo{
			Name: name,
			Pc:   pc,
		}
		if apiteam == "a" {
			algo.ApiUrl = api.API_URL_A
		} else {
			algo.ApiUrl = api.API_URL_B
		}
		if algo.Pc.ID == "" {
			if ok, _ := algo.GetInfosProgramme(); !ok {
				err = errors.New("erreur get infos programme")
				return
			}
		}
		algo.ID = algo.Psi.Programme.ID
	}
	return algo, err
}
func NewAlgoBlueTeam(name string, apiteam string) (algo *Algo, err error) {
	tools.Title(fmt.Sprintf("chargement programme [%s]", name))
	psi, pc, err := _LoadProgrammeBlueTeam(name, apiteam)
	algo = &Algo{
		Name:   name,
		Psi:    psi,
		Pc:     pc,
		ApiUrl: apiteam,
	}
	if apiteam == "a" {
		algo.ApiUrl = api.API_URL_A
	} else {
		algo.ApiUrl = api.API_URL_B
	}
	if algo.Psi.Programme.ID == "" {
		if ok, _ := algo.GetInfosProgramme(); !ok {
			err = errors.New("erreur get infos programme")
			return
		}
	}
	algo.ID = algo.Psi.Programme.ID
	return algo, err
}
func (a *Algo) LoadProgramme() (ok bool, err error) {
	reqBodyBytes := new(bytes.Buffer)
	json.NewEncoder(reqBodyBytes).Encode(a.Pc)
	res, statusCode, _ := api.RequestApi(
		"POST",
		fmt.Sprintf("%s/%s", a.ApiUrl, api.ROUTE_LOAD_PROGRAMME),
		reqBodyBytes.Bytes(),
	)
	if statusCode == http.StatusCreated || statusCode == http.StatusOK {
		_ = json.Unmarshal(res, a.Psi)
		return true, err
	} else {
		err = errors.New("erreur chargement programme")
	}
	return false, err
}
func (a *Algo) GetInfosProgramme() (ok bool, err error) {
	//tools.Title(fmt.Sprintf("infos programme [%s]", a.Name))
	res, statusCode, err := api.RequestApi(
		"GET",
		fmt.Sprintf("%s/%s/%s/%s", a.ApiUrl, api.ROUTE_STATUS_PROGRAMME, a.Pc.ID, a.Pc.SecretID),
		nil,
	)
	a.StatusCode = statusCode
	if err != nil || statusCode != http.StatusOK {
		return false, err
	}
	a.Psi = structure.ProgrammeStatusInfos{}
	err = json.Unmarshal(res, &a.Psi)
	return true, err
}
func (a *Algo) NavigationStop() (ok bool, err error) {
	res, statusCode, err := api.RequestApi(
		"GET",
		fmt.Sprintf("%s/%s/%s/%s", a.ApiUrl, api.ROUTE_NAVIGATION_PROGRAMME_STOP, a.Pc.ID, a.Pc.SecretID),
		nil,
	)
	a.StatusCode = statusCode
	if err != nil || statusCode != http.StatusOK {
		return false, err
	}
	a.Psi = structure.ProgrammeStatusInfos{}
	err = json.Unmarshal(res, &a.Psi)
	return true, err
}
func (a *Algo) ExplorationStop() (ok bool, err error) {
	res, statusCode, err := api.RequestApi(
		"GET",
		fmt.Sprintf("%s/%s/%s/%s", a.ApiUrl, api.ROUTE_EXPLORATION_PROGRAMME_STOP, a.Pc.ID, a.Pc.SecretID),
		nil,
	)
	a.StatusCode = statusCode
	if err != nil || statusCode != http.StatusOK {
		return false, err
	}
	a.Psi = structure.ProgrammeStatusInfos{}
	err = json.Unmarshal(res, &a.Psi)
	return true, err
}
func (a *Algo) Move(secteurID string, zoneID string) (ok bool, err error) {
	tools.Title(fmt.Sprintf("Programme [%s] Move to S%s-Z%s", a.Name, secteurID, zoneID))
	res, statusCode, err := api.RequestApi(
		"GET",
		fmt.Sprintf("%s/%s/%s/%s/%s/%s", a.ApiUrl, api.ROUTE_MOVE_PROGRAMME, a.Pc.ID, a.Pc.SecretID, secteurID, zoneID),
		nil,
	)
	if err != nil || statusCode != http.StatusOK {
		return false, err
	}
	a.Psi = structure.ProgrammeStatusInfos{}
	err = json.Unmarshal(res, &a.Psi)
	return true, err
}
func (a *Algo) QuickMove(secteurID string, zoneID string) (ok bool, err error) {
	tools.Title(fmt.Sprintf("Programme [%s] Move to S%s-Z%s", a.Name, secteurID, zoneID))
	res, statusCode, err := api.RequestApi(
		"GET",
		fmt.Sprintf("%s/%s/%s/%s/%s/%s", a.ApiUrl, api.ROUTE_QUICK_MOVE_PROGRAMME, a.Pc.ID, a.Pc.SecretID, secteurID, zoneID),
		nil,
	)
	if err != nil || statusCode != http.StatusOK {
		return false, err
	}
	a.Psi = structure.ProgrammeStatusInfos{}
	err = json.Unmarshal(res, &a.Psi)
	return true, err
}
func (a *Algo) EstimateMove(secteurID string, zoneID string) (data structure.MoveEstimateData, err error) {
	tools.Title(fmt.Sprintf("Programme [%s] Estimate Move to S%s-Z%s", a.Name, secteurID, zoneID))
	res, statusCode, err := api.RequestApi(
		"GET",
		fmt.Sprintf("%s/%s/%s/%s/%s/%s", a.ApiUrl, api.ROUTE_ESTIMATE_MOVE_PROGRAMME, a.Pc.ID, a.Pc.SecretID, secteurID, zoneID),
		nil,
	)
	if err != nil || statusCode != http.StatusOK {
		return data, err
	}
	err = json.Unmarshal(res, &data)
	return data, err
}
func (a *Algo) StopMove() (ok bool, err error) {
	tools.Title(fmt.Sprintf("Programme [%s] stop move", a.Name))
	res, statusCode, err := api.RequestApi(
		"GET",
		fmt.Sprintf("%s/%s/%s/%s", a.ApiUrl, api.ROUTE_STOP_MOVE_PROGRAMME, a.Pc.ID, a.Pc.SecretID),
		nil,
	)
	if err != nil || statusCode != http.StatusOK {
		return false, err
	}
	a.Psi = structure.ProgrammeStatusInfos{}
	err = json.Unmarshal(res, &a.Psi)
	return true, err
}
func (a *Algo) Scan() (ok bool, res []byte, err error) {
	res, statusCode, err := api.RequestApi(
		"GET",
		fmt.Sprintf("%s/%s/%s/%s", a.ApiUrl, api.ROUTE_SCAN_PROGRAMME, a.Pc.ID, a.Pc.SecretID),
		nil,
	)
	if err != nil || statusCode != http.StatusOK {
		return false, res, err
	}
	return true, res, err
}
func (a *Algo) Explore(celluleID int) (ok bool, res []byte, err error) {
	tools.Title(fmt.Sprintf("Programme [%s] explore cellule [%d]", a.Name, celluleID))
	res, statusCode, err := api.RequestApi(
		"GET",
		fmt.Sprintf("%s/%s/%s/%s/%d", a.ApiUrl, api.ROUTE_EXPLORE_PROGRAMME, a.Pc.ID, a.Pc.SecretID, celluleID),
		nil,
	)
	if err != nil || statusCode != http.StatusOK {
		return false, res, err
	}
	return true, res, err
}
func (a *Algo) DestroyZone(celluleID int, energy int) (ok bool, res []byte, err error) {
	title := aurora.Red("--- Destroy zone")
	tools.Title(fmt.Sprintf("\t%s >>> cellule [%d]", title, celluleID))
	res, statusCode, err := api.RequestApi(
		"GET",
		fmt.Sprintf("%s/%s/%s/%s/%d/%d", a.ApiUrl, api.ROUTE_DESTROY_ZONE, a.Pc.ID, a.Pc.SecretID, celluleID, energy),
		nil,
	)
	if err != nil || statusCode != http.StatusOK {
		return false, res, err
	}
	a.Psi = structure.ProgrammeStatusInfos{}
	err = json.Unmarshal(res, &a.Psi)
	return true, res, err
}
func (a *Algo) Destroy(celluleID int, targetID string, energy int) (ok bool, res []byte, err error) {
	title := aurora.Red("--- Destroy programme")
	tools.Title(fmt.Sprintf("\t%s >>> [%s] cellule [%d]", title, aurora.Cyan(targetID), celluleID))
	res, statusCode, err := api.RequestApi(
		"GET",
		fmt.Sprintf("%s/%s/%s/%s/%d/%s/%d", a.ApiUrl, api.ROUTE_DESTROY_PROGRAMME, a.Pc.ID, a.Pc.SecretID, celluleID, targetID, energy),
		nil,
	)
	if err != nil || statusCode != http.StatusOK {
		return false, res, err
	}
	a.Psi = structure.ProgrammeStatusInfos{}
	err = json.Unmarshal(res, &a.Psi)
	return true, res, err
}
func (a *Algo) Rebuild(celluleID int, targetID string, energy int) (ok bool, res []byte, err error) {
	title := aurora.Blue("+++ Rebuild programme")
	tools.Title(fmt.Sprintf("\t%s >>> [%s] cellule [%d]", title, aurora.Cyan(targetID), celluleID))
	res, statusCode, err := api.RequestApi(
		"GET",
		fmt.Sprintf("%s/%s/%s/%s/%d/%s/%d", a.ApiUrl, api.ROUTE_REBUILD_PROGRAMME, a.Pc.ID, a.Pc.SecretID, celluleID, targetID, energy),
		nil,
	)
	if err != nil || statusCode != http.StatusOK {
		return false, res, err
	}
	a.Psi = structure.ProgrammeStatusInfos{}
	err = json.Unmarshal(res, &a.Psi)
	return true, res, err
}
func (a *Algo) GetStatusGrid() (err error) {
	res, statusCode, err := api.RequestApi(
		"GET",
		fmt.Sprintf("%s/%s", a.ApiUrl, api.ROUTE_STATUS_GRID),
		nil,
	)
	if err != nil {
		err = errors.New(fmt.Sprintf("status code [%d] - [%s]", statusCode, err.Error()))
	} else {
		a.InfosGrid = structure.GridInfos{}
		err = json.Unmarshal(res, &a.InfosGrid)
	}
	return
}
func (a *Algo) PrintInfo(printGrid bool) {
	err := a.GetStatusGrid()
	if err != nil {
		tools.Fail(err.Error())
	}
	tools.PrintProgramme(a.Psi)
	tools.PrintInfosGrille(a.InfosGrid)
	if printGrid {
		tools.PrintGridPosition(a.Psi.Programme, a.InfosGrid.Taille)
	}
}
func (a *Algo) GetProgramme() (ok bool, programmes []string) {
	if okZI, zoneInfos := a.GetZoneinfos(); okZI {
		for _, programme := range zoneInfos.Programmes {
			if programme.Status {
				programmes = append(programmes, programme.Name)
			}
		}
		return true, programmes
	}
	return false, programmes
}
func (a *Algo) GetZoneinfos() (ok bool, zoneInfos structure.ZoneInfos) {
	if scanOK, res, _ := a.Scan(); !scanOK {
		return scanOK, zoneInfos
	} else {
		err := json.Unmarshal(res, &zoneInfos)
		if err != nil {
			return false, zoneInfos
		}
	}
	return true, zoneInfos
}
func (a *Algo) SearchFlag(cellules []structure.CelluleInfos) (flagFound bool) {
	for _, cellule := range cellules {
		if cellule.Status {
			//tools.Success(fmt.Sprintf("zone [%d] - cellule [%d] etat [%t] - data presente ou etat true", zoneInfos.ID, cellule.ID, cellule.Status))
			if exploreOK, exploreRes, _ := a.Explore(cellule.ID); !exploreOK {
				jsonPretty, _ := tools.PrettyString(exploreRes)
				fmt.Println(jsonPretty)
				tools.Fail("erreur explore")
			} else {
				var datas map[int]structure.CelluleData
				err := json.Unmarshal(exploreRes, &datas)
				if err != nil {
					tools.Fail(err.Error())
				} else {
					for _, data := range datas {
						if data.IsFlag {
							tools.Success(fmt.Sprintf("FLAG TROUVER - Cellule [%d] - index [%d]", cellule.ID, data.ID))
							a.CaptureCellData(cellule.ID, data.ID)
							return true
						}
					}
				}
			}
		} else {
			//tools.Warning(fmt.Sprintf("cellule [%d] etat [%t] - aucune data ou etat false", cellule.ID, cellule.Status))
		}
	}
	return
}
func (a *Algo) SearchEnergy(cellules []structure.CelluleInfos) (index int) {
	for _, cellule := range cellules {
		if cellule.Trapped {
			title := aurora.Red("--- CELLULE DANGER")
			tools.Title(fmt.Sprintf(
				"\t%s >>> [%d][%d] cellule [%d]",
				title,
				a.Psi.Programme.Position.SecteurID,
				a.Psi.Programme.Position.ZoneID,
				cellule.ID,
			))
		} else {
			if cellule.Status {
				if exploreOK, exploreRes, _ := a.Explore(cellule.ID); !exploreOK {
					jsonPretty, _ := tools.PrettyString(exploreRes)
					fmt.Println(jsonPretty)
					tools.Fail("erreur explore")
				} else {
					var datas map[int]structure.CelluleData
					err := json.Unmarshal(exploreRes, &datas)
					if err != nil {
						tools.Fail(err.Error())
					} else {
						for _, data := range datas {
							if data.Energy > 0 {
								a.CaptureCellEnergy(cellule.ID, data.ID)
							}
						}
					}
				}
			} else {
				//tools.Warning(fmt.Sprintf("cellule [%d] etat [%t] - aucune data ou etat false", cellule.ID, cellule.Status))
			}
		}
	}
	return
}
func (a *Algo) CaptureCellData(celluleID int, index int) (ok bool, err error) {
	res, statusCode, err := api.RequestApi(
		"GET",
		fmt.Sprintf("%s/%s/%s/%s/%d/%d", a.ApiUrl, api.ROUTE_CAPTURE_CELL_DATA, a.Pc.ID, a.Pc.SecretID, celluleID, index),
		nil,
	)
	if err != nil {
		tools.Fail(fmt.Sprintf("status code [%d] - [%s]", statusCode, err.Error()))
	} else {
		if err != nil || statusCode != http.StatusOK {
			return false, err
		}
		a.Psi = structure.ProgrammeStatusInfos{}
		err = json.Unmarshal(res, &a.Psi)
	}
	return
}
func (a *Algo) CaptureTargetData(celluleID int, targetID string) (ok bool, err error) {
	res, statusCode, err := api.RequestApi(
		"GET",
		fmt.Sprintf("%s/%s/%s/%s/%d/%s", a.ApiUrl, api.ROUTE_CAPTURE_TARGET_DATA, a.Pc.ID, a.Pc.SecretID, celluleID, targetID),
		nil,
	)
	if err != nil {
		tools.Fail(fmt.Sprintf("status code [%d] - [%s]", statusCode, err.Error()))
	} else {
		if err != nil || statusCode != http.StatusOK {
			return false, err
		}
		a.Psi = structure.ProgrammeStatusInfos{}
		err = json.Unmarshal(res, &a.Psi)
	}
	return
}
func (a *Algo) CaptureTargetEnergy(celluleID int, targetID string) (ok bool, err error) {
	res, statusCode, err := api.RequestApi(
		"GET",
		fmt.Sprintf("%s/%s/%s/%s/%d/%s", a.ApiUrl, api.ROUTE_CAPTURE_TARGET_ENERGY, a.Pc.ID, a.Pc.SecretID, celluleID, targetID),
		nil,
	)
	if err != nil {
		tools.Fail(fmt.Sprintf("status code [%d] - [%s]", statusCode, err.Error()))
	} else {
		if err != nil || statusCode != http.StatusOK {
			return false, err
		}
		a.Psi = structure.ProgrammeStatusInfos{}
		err = json.Unmarshal(res, &a.Psi)
	}
	return
}
func (a *Algo) CaptureCellEnergy(celluleID int, index int) (ok bool, err error) {
	res, statusCode, err := api.RequestApi(
		"GET",
		fmt.Sprintf("%s/%s/%s/%s/%d/%d", a.ApiUrl, api.ROUTE_CAPTURE_CELL_ENERGY, a.Pc.ID, a.Pc.SecretID, celluleID, index),
		nil,
	)
	if err != nil {
		tools.Fail(fmt.Sprintf("status code [%d] - [%s]", statusCode, err.Error()))
	} else {
		if err != nil || statusCode != http.StatusOK {
			return false, err
		}
		a.Psi = structure.ProgrammeStatusInfos{}
		err = json.Unmarshal(res, &a.Psi)
	}
	return
}
func (a *Algo) Equilibrium() (ok bool, err error) {
	res, statusCode, err := api.RequestApi(
		"GET",
		fmt.Sprintf("%s/%s/%s/%s", a.ApiUrl, api.ROUTE_EQUILIBRiUM, a.Pc.ID, a.Pc.SecretID),
		nil,
	)
	if err != nil {
		tools.Fail(fmt.Sprintf("status code [%d] - [%s]", statusCode, err.Error()))
	} else {
		if err != nil || statusCode != http.StatusOK {
			return false, err
		}
		a.Psi = structure.ProgrammeStatusInfos{}
		err = json.Unmarshal(res, &a.Psi)
	}
	return
}
func (a *Algo) PushFlag() (ok bool, err error) {
	res, statusCode, err := api.RequestApi(
		"GET",
		fmt.Sprintf("%s/%s/%s/%s", a.ApiUrl, api.ROUTE_PUSH_FLAG, a.Pc.ID, a.Pc.SecretID),
		nil,
	)
	if err != nil {
		tools.Fail(fmt.Sprintf("status code [%d] - [%s]", statusCode, err.Error()))
	} else {
		if err != nil || statusCode != http.StatusOK {
			tools.Fail("backup FAIL")
			return false, err
		}
		err = json.Unmarshal(res, &a.Pc)
		tools.CreateJsonFile(fmt.Sprintf("%s.json", a.Name), a.Pc)
		tools.Success("backup OK")
		ok = true
	}
	return
}
func (a *Algo) ShellCode() (ok bool, data []structure.ShellcodeData, err error) {
	res, statusCode, err := api.RequestApi(
		"GET",
		fmt.Sprintf("%s/%s/%s/%s", a.ApiUrl, api.ROUTE_ZONE_SHELLCODE, a.Pc.ID, a.Pc.SecretID),
		nil,
	)
	if err != nil {
		tools.Fail(fmt.Sprintf("status code [%d] - [%s]", statusCode, err.Error()))
	} else {
		if err != nil || statusCode != http.StatusOK {
			tools.Fail("erreur")
			return false, data, err
		}
		err = json.Unmarshal(res, &data)
		ok = true
	}
	return
}
func (a *Algo) ActiveShellCode(targetID string, shellcode string) (ok bool, err error) {
	_, statusCode, err := api.RequestApi(
		"GET",
		fmt.Sprintf("%s/%s/%s/%s/%s/%s", a.ApiUrl, api.ROUTE_ACTIVE_SHELLCODE, a.Pc.ID, a.Pc.SecretID, targetID, shellcode),
		nil,
	)
	if err != nil {
		tools.Fail(fmt.Sprintf("status code [%d] - [%s]", statusCode, err.Error()))
	} else {
		if err != nil || statusCode != http.StatusOK {
			tools.Fail("backup FAIL")
			return false, err
		}
		tools.Success("OK")
		ok = true
	}
	return
}
func (a *Algo) ActiveCaptureFlag(Flag string) (ok bool, err error) {
	_, statusCode, err := api.RequestApi(
		"GET",
		fmt.Sprintf("%s/%s/%s/%s/%s", a.ApiUrl, api.ROUTE_ACTIVE_CAPTURE_FLAG, a.Pc.ID, a.Pc.SecretID, Flag),
		nil,
	)
	if err != nil {
		tools.Fail(fmt.Sprintf("status code [%d] - [%s]", statusCode, err.Error()))
	} else {
		if err != nil || statusCode != http.StatusOK {
			tools.Fail("backup FAIL")
			return false, err
		}
		tools.Success("OK")
		ok = true
	}
	return
}
