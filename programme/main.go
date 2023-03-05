package programme

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/greg198584/client-ag3/algo"
	"github.com/greg198584/client-ag3/api"
	"github.com/greg198584/client-ag3/cia_engine"
	"github.com/greg198584/client-ag3/structure"
	"github.com/greg198584/client-ag3/tools"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func _IsExistFile(name string) bool {
	filename := fmt.Sprintf("%s.json", name)
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return false
	} else {
		return true
	}
}
func _CreateProgramme(name string, apiteam string) (programme structure.ProgrammeContainer, err error) {
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
		return
	}
	if statusCode == http.StatusCreated {
		err = json.Unmarshal(res, &programme)
		tools.CreateJsonFile(fmt.Sprintf("%s.json", name), programme)
	} else {
		err = errors.New("erreur creation programme")
	}
	return
}
func New(name string, apiteam string) {
	tools.Title(fmt.Sprintf("création programme [%s]", name))
	if _IsExistFile(name) == false {

		programmeContainer, err := _CreateProgramme(name, apiteam)
		if err != nil {
			tools.Fail(err.Error())
		} else {
			tools.Success(fmt.Sprintf("programme ajouter ID = [%s]", programmeContainer.ID))
			tools.Info(fmt.Sprintf("programme info"))
			Info(&programmeContainer)
		}
	} else {
		tools.Warning(fmt.Sprintf("programme file exist"))
	}
}
func GetApiUrl(apiTeam string) (url string) {
	if apiTeam == "a" {
		url = api.API_URL_A
	} else {
		url = api.API_URL_B
	}
	return
}
func Info(pc *structure.ProgrammeContainer) {
	reqBodyBytes := new(bytes.Buffer)
	json.NewEncoder(reqBodyBytes).Encode(pc.Programme)
	jsonPretty, _ := tools.PrettyString(reqBodyBytes.Bytes())
	fmt.Println(jsonPretty)
}
func Load(name string, apiteam string, team string) {
	var err error
	var current *algo.Algo
	tools.Title(fmt.Sprintf("chargement programme [%s]", name))
	if team == "blue" {
		current, err = algo.NewAlgoBlueTeam(name, GetApiUrl(apiteam))
		if err != nil {
			tools.Fail(err.Error())
			return
		}
	} else {
		current, err = algo.NewAlgo(name, GetApiUrl(apiteam))
		current.LoadProgramme()
		current.GetInfosProgramme()
		if err != nil {
			tools.Fail(err.Error())
			return
		}
	}
	current.PrintInfo(true)
}
func Scan(name string, apiteam string) {
	tools.Title(fmt.Sprintf("Programme [%s] scan", name))
	current, err := algo.NewAlgo(name, GetApiUrl(apiteam))
	if err != nil {
		//panic(err)
	}
	_, res, err := current.Scan()
	if err != nil {
		tools.Fail(err.Error())
	} else {
		var zoneInfos structure.ZoneInfos
		err = json.Unmarshal(res, &zoneInfos)
		if err != nil {
			tools.Fail(err.Error())
		} else {
			tools.PrintZoneInfos(zoneInfos)
		}
	}
}
func Explore(name string, apiteam string, celluleID string) {
	tools.Title(fmt.Sprintf("Programme [%s] explore cellule [%s]", name, celluleID))
	current, err := algo.NewAlgo(name, GetApiUrl(apiteam))
	if err != nil {
		//panic(err)
	}
	celluleIdInt, err := strconv.Atoi(celluleID)
	_, res, err := current.Explore(celluleIdInt)
	if err != nil {
		tools.Fail(err.Error())
	} else {
		var celluleData map[int]structure.CelluleData
		err = json.Unmarshal(res, &celluleData)
		if err != nil {
			tools.Fail(err.Error())
		} else {
			tools.PrintExplore(celluleID, celluleData)
		}
	}
}
func Destroy(name string, apiteam string, celluleID int, targetID string, energy int) {
	tools.Title(fmt.Sprintf(
		"Programme [%s] destroy -> [%s] cellule [%s] energy [%s]",
		name,
		celluleID,
		targetID,
		algo.ENERGY_MAX_ATTACK,
	))
	current, err := algo.NewAlgo(name, GetApiUrl(apiteam))
	if err != nil {
		//panic(err)
	}
	current.Destroy(celluleID, targetID, energy)
	current.PrintInfo(false)
	return
}
func Rebuild(name string, apiteam string, celluleID int, targetID string, energy int) {
	tools.Title(fmt.Sprintf(
		"Programme [%s] rebuild -> [%s] cellule [%s] energy [%s]",
		name,
		celluleID,
		targetID,
		algo.ENERGY_MAX_ATTACK,
	))
	current, err := algo.NewAlgo(name, GetApiUrl(apiteam))
	if err != nil {
		//panic(err)
	}
	current.Rebuild(celluleID, targetID, energy)
	current.PrintInfo(false)
	return
}
func GetStatusGrid(apiteam string, zoneActif bool) {
	tools.Title(fmt.Sprintf("Status grid"))
	res, statusCode, err := api.RequestApi(
		"GET",
		fmt.Sprintf("%s/%s", GetApiUrl(apiteam), api.ROUTE_STATUS_GRID),
		nil,
	)
	if err != nil {
		tools.Fail(fmt.Sprintf("status code [%d] - [%s]", statusCode, err.Error()))
	} else {
		var infos structure.GridInfos
		err = json.Unmarshal(res, &infos)
		if err != nil {
			tools.Fail(err.Error())
		} else {
			var zonesList []structure.ZonesGrid
			if zoneActif {
				for _, zone := range infos.Zones {
					if zone.Status {
						zonesList = append(zonesList, zone)
					}
				}
			} else {
				zonesList = infos.Zones
			}
			tools.PrintZoneActif(zonesList)
			tools.PrintInfosGrille(infos)
		}
	}
	return
}
func GetInfoProgramme(name string, apiteam string, printPosition bool) {
	tools.Title(fmt.Sprintf("infos programme"))
	current, err := algo.NewAlgo(name, GetApiUrl(apiteam))
	if err != nil {
		//panic(err)
	}
	current.GetInfosProgramme()
	current.PrintInfo(printPosition)
}
func Navigation(name string, apiteam string) {
	tools.Title(fmt.Sprintf("stop mode navigation programme"))
	current, err := algo.NewAlgo(name, GetApiUrl(apiteam))
	if err != nil {
		//panic(err)
	}
	current.NavigationStop()
	current.PrintInfo(false)
}
func ExplorationStop(name string, apiteam string) {
	tools.Title(fmt.Sprintf("stop exploration"))
	current, err := algo.NewAlgo(name, GetApiUrl(apiteam))
	if err != nil {
		//panic(err)
	}
	current.ExplorationStop()
	current.PrintInfo(false)
}
func CaptureTargetData(name string, apiteam string, celluleID int, targetID string) {
	tools.Title(fmt.Sprintf("[%s] Capture data target [%s] - cellule [%s]", name, targetID, celluleID))
	current, err := algo.NewAlgo(name, GetApiUrl(apiteam))
	if err != nil {
		//panic(err)
	}
	current.CaptureTargetData(celluleID, targetID)
	current.PrintInfo(false)
	return
}
func CaptureCellData(name string, apiteam string, celluleID int, index string) {
	tools.Title(fmt.Sprintf("[%s] Capture data cellule [%d] - index [%d]", name, celluleID, index))
	index_split := strings.Split(index, "-")
	current, _ := algo.NewAlgo(name, GetApiUrl(apiteam))
	if len(index_split) > 1 {
		id, _ := strconv.Atoi(index_split[0])
		count, _ := strconv.Atoi(index_split[1])
		for id < count+1 {
			current.CaptureCellData(celluleID, id)
			id++
		}
	} else {
		id, _ := strconv.Atoi(index)
		current.CaptureCellData(celluleID, id)
	}
	current.PrintInfo(false)
	return
}
func CaptureTargetEnergy(name string, apiteam string, celluleID int, targetID string) {
	tools.Title(fmt.Sprintf("[%s] Capture energy target [%s] - cellule [%s]", name, targetID, celluleID))
	current, err := algo.NewAlgo(name, GetApiUrl(apiteam))
	if err != nil {
		//panic(err)
	}
	current.CaptureTargetEnergy(celluleID, targetID)
	current.PrintInfo(false)
	return
}
func CaptureCellEnergy(name string, apiteam string, celluleID int, index string) {
	tools.Title(fmt.Sprintf("[%s] Capture energy cellule [%s] - index [%d]", name, celluleID, index))
	index_split := strings.Split(index, "-")
	fmt.Printf("index_split = [%v]\n", index_split)
	current, _ := algo.NewAlgo(name, GetApiUrl(apiteam))
	if len(index_split) > 1 {
		id, _ := strconv.Atoi(index_split[0])
		count, _ := strconv.Atoi(index_split[1])
		for id < count+1 {
			current.CaptureCellEnergy(celluleID, id)
			id++
		}
	} else {
		id, _ := strconv.Atoi(index)
		current.CaptureCellEnergy(celluleID, id)
	}
	current.PrintInfo(false)
	return
}
func Equilibrium(name string, apiteam string) {
	tools.Title(fmt.Sprintf("Equilibrium energy programme [%s]", name))
	current, err := algo.NewAlgo(name, GetApiUrl(apiteam))
	if err != nil {
		//panic(err)
	}
	current.Equilibrium()
	current.PrintInfo(false)
}
func PushFlag(name string, apiteam string) {
	tools.Title(fmt.Sprintf("Push flag - programme [%s]", name))
	current, err := algo.NewAlgo(name, GetApiUrl(apiteam))
	if err != nil {
		//panic(err)
	}
	current.PushFlag()
	current.GetInfosProgramme()
	current.PrintInfo(false)
}
func DestroyZone(name string, apiteam string, celluleID int, energy int, all bool) {
	current, err := algo.NewAlgo(name, GetApiUrl(apiteam))
	if err != nil {
		//panic(err)
	}
	ok, zoneInfos := current.GetZoneinfos()
	if ok && zoneInfos.Status {
		if all {
			for _, cellule := range zoneInfos.Cellules {
				current.DestroyZone(cellule.ID, energy)
			}
		} else {
			current.DestroyZone(celluleID, energy)
		}
		tools.PrintZoneInfos(zoneInfos)
	}
	_, zoneInfos = current.GetZoneinfos()
	tools.PrintZoneInfos(zoneInfos)
}

func Monitoring(name string, apiteam string, printGrid bool) {
	current, err := algo.NewAlgo(name, GetApiUrl(apiteam))
	if err != nil {
		//panic(err)
	}
	for {
		time.Sleep(algo.TIME_MILLISECONDE * time.Millisecond)
		current.GetInfosProgramme()
		current.PrintInfo(printGrid)
	}
}
func GetCelluleLog(name string, apiteam string, celluleID string) {
	tools.Title(fmt.Sprintf("GET LOG cellule [%s] - programme [%s]", celluleID, name))
	current, err := algo.NewAlgo(name, GetApiUrl(apiteam))
	if err != nil {
		//panic(err)
	}
	res, statusCode, err := api.RequestApi(
		"GET",
		fmt.Sprintf("%s/%s/%s/%s/%s", current.ApiUrl, api.ROUTE_GET_CELLULE_LOG, current.Pc.ID, current.Pc.SecretID, celluleID),
		nil,
	)
	if err != nil {
		tools.Fail(fmt.Sprintf("status code [%d] - [%s]", statusCode, err.Error()))
	} else {
		var celluleLogs map[int]structure.CelluleLog
		err = json.Unmarshal(res, &celluleLogs)
		if err != nil {
			tools.Fail(err.Error())
		} else {
			tools.PrintCelluleLogs(celluleLogs)
		}
	}
	return
}
func CleanCelluleLog(name string, apiteam string, celluleID string) {
	tools.Title(fmt.Sprintf("CLEAN LOG cellule [%s] - programme [%s]", celluleID, name))
	current, err := algo.NewAlgo(name, GetApiUrl(apiteam))
	if err != nil {
		//panic(err)
	}
	_, statusCode, err := api.RequestApi(
		"GET",
		fmt.Sprintf("%s/%s/%s/%s/%s", current.ApiUrl, api.ROUTE_CLEAN_CELLULE_LOG, current.Pc.ID, current.Pc.SecretID, celluleID),
		nil,
	)
	if err != nil {
		tools.Fail(fmt.Sprintf("status code [%d] - [%s]", statusCode, err.Error()))
	} else {
		tools.Success("clean cellule")
	}
	return
}
func MovePosition(name string, apiteam string, secteurID string, zoneID string) {
	current, err := algo.NewAlgo(name, GetApiUrl(apiteam))
	if err != nil {
		//panic(err)
	}
	current.Move(secteurID, zoneID)
	current.PrintInfo(true)
}
func QuickMovePosition(name string, apiteam string, secteurID string, zoneID string) {
	current, err := algo.NewAlgo(name, GetApiUrl(apiteam))
	if err != nil {
		//panic(err)
	}
	current.QuickMove(secteurID, zoneID)
	current.PrintInfo(true)
}
func EstimateMove(name string, apiteam string, secteurID string, zoneID string) {
	current, err := algo.NewAlgo(name, GetApiUrl(apiteam))
	if err != nil {
		//panic(err)
	}
	data, _ := current.EstimateMove(secteurID, zoneID)
	var header = []string{"Secteur_ID", "Zone_ID", "Distance", "Estimation", "Cout_Energy", "Cout_Iteration"}
	var dataTab [][]string

	dataTab = append(dataTab, []string{
		fmt.Sprintf("%d", data.SecteurID),
		fmt.Sprintf("%d", data.ZoneID),
		fmt.Sprintf("%d", data.Distance),
		fmt.Sprintf("%s", data.TempEstimate),
		fmt.Sprintf("%d", data.CoutEnergy),
		fmt.Sprintf("%d", data.CoutIteration),
	})
	tools.PrintColorTable(header, dataTab, "<---[ Estimation temp de deplacement ]--->")
}
func StopMove(name string, apiteam string) {
	current, err := algo.NewAlgo(name, GetApiUrl(apiteam))
	if err != nil {
		//panic(err)
	}
	current.StopMove()
	current.PrintInfo(true)
}
func ShellCode(name string, apiteam string) {
	current, err := algo.NewAlgo(name, GetApiUrl(apiteam))
	if err != nil {
		//panic(err)
	}
	ok, data, _ := current.ShellCode()
	if ok {
		tools.PrintShellCodeData(data)
	}
}
func ActiveShellCode(name string, apiteam string, targetID string, ShellCode string) {
	current, err := algo.NewAlgo(name, GetApiUrl(apiteam))
	if err != nil {
		tools.Fail(err.Error())
	} else {
		current.ActiveShellCode(targetID, ShellCode)
	}
}
func ActiveCaptureFlag(name string, apiteam string, Flag string) {
	current, err := algo.NewAlgo(name, GetApiUrl(apiteam))
	if err != nil {
		tools.Fail(err.Error())
	} else {
		current.ActiveCaptureFlag(Flag)
	}
}
func RunCIA(name string) {
	current, err := cia_engine.New(name)
	if err != nil {
		tools.Fail(err.Error())
	} else {
		err = current.Run()
		if err != nil {
			tools.Fail(err.Error())
		}
	}
}
