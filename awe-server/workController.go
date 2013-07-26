package main

import (
	"github.com/MG-RAST/AWE/core"
	e "github.com/MG-RAST/AWE/errors"
	. "github.com/MG-RAST/AWE/logger"
	"github.com/jaredwilkening/goweb"
	"io/ioutil"
	"net/http"
	"strings"
)

type WorkController struct{}

// GET: /work/{id}
// get a workunit by id, read-only
func (cr *WorkController) Read(id string, cx *goweb.Context) {
	LogRequest(cx.Request)

	// Load workunit by id
	workunit, err := queueMgr.GetWorkById(id)

	if err != nil {
		if err.Error() != e.QueueEmpty {
			Log.Error("Err@work_Read:QueueMgr.GetWorkById(): " + err.Error())
		}
		cx.RespondWithErrorMessage(err.Error(), http.StatusBadRequest)
		return
	}
	// Base case respond with workunit in json
	cx.RespondWithData(workunit)
	return
}

// GET: /work
// checkout a workunit with earliest submission time
// to-do: to support more options for workunit checkout
func (cr *WorkController) ReadMany(cx *goweb.Context) {

	// Gather query params
	query := &Query{list: cx.Request.URL.Query()}

	if !query.Has("client") { //view workunits
		var workunits []*core.Workunit
		if query.Has("state") {
			workunits = queueMgr.ShowWorkunits(query.Value("state"))
		} else {
			workunits = queueMgr.ShowWorkunits("")
		}
		cx.RespondWithData(workunits)
		return
	}

	//checkout a workunit in FCFS order
	clientid := query.Value("client")
	workunits, err := queueMgr.CheckoutWorkunits("FCFS", clientid, 1)

	if err != nil {
		if err.Error() != e.QueueEmpty && err.Error() != e.NoEligibleWorkunitFound {
			Log.Error("Err@work_ReadMany:QueueMgr.GetWorkByFCFS(): " + err.Error())
		}
		cx.RespondWithErrorMessage(err.Error(), http.StatusBadRequest)
		return
	}

	//log access info only when the queue is not empty, save some log
	LogRequest(cx.Request)

	//log event about workunit checkout (WO)
	workids := []string{}
	for _, work := range workunits {
		workids = append(workids, work.Id)
	}

	Log.Event(EVENT_WORK_CHECKOUT,
		"workids="+strings.Join(workids, ","),
		"clientid="+clientid)

	// Base case respond with node in json
	cx.RespondWithData(workunits[0])
	return
}

// PUT: /work/{id} -> status update
func (cr *WorkController) Update(id string, cx *goweb.Context) {
	// Log Request and check for Auth
	LogRequest(cx.Request)
	// Gather query params
	query := &Query{list: cx.Request.URL.Query()}

	if query.Has("status") && query.Has("client") { //notify execution result: "done" or "fail"
		notice := core.Notice{WorkId: id, Status: query.Value("status"), ClientId: query.Value("client"), Notes: ""}
		if query.Has("report") { // if "report" is specified in query, parse performance statistics or errlog
			if _, files, err := ParseMultipartForm(cx.Request); err == nil {
				if _, ok := files["perf"]; ok {
					queueMgr.FinalizeWorkPerf(id, files["perf"].Path)
				}
				if _, ok := files["notes"]; ok {
					if notes, err := ioutil.ReadFile(files["notes"].Path); err == nil {
						notice.Notes = string(notes)
					}
				}
			}
		}
		queueMgr.NotifyWorkStatus(notice)
	}
	cx.RespondWithData("ok")
	return
}
