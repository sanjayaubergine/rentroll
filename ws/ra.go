package ws

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"rentroll/rlib"
	"strings"
	"time"
)

// RentalAgr is a structure specifically for the Web Services interface. It will be
// automatically populated from an rlib.RentalAgreement struct. Records of this type
// are returned by the search handler
type RentalAgr struct {
	Recid                  int64             `json:"recid"` // this is to support the w2ui form
	RAID                   int64             // internal unique id
	RATID                  int64             // reference to Occupancy Master Agreement
	BID                    rlib.XJSONBud     // Business (so that we can process by Business)
	NLID                   int64             // Note ID
	AgreementStart         rlib.JSONTime     // start date for rental agreement contract
	AgreementStop          rlib.JSONTime     // stop date for rental agreement contract
	PossessionStart        rlib.JSONTime     // start date for Occupancy
	PossessionStop         rlib.JSONTime     // stop date for Occupancy
	RentStart              rlib.JSONTime     // start date for Rent
	RentStop               rlib.JSONTime     // stop date for Rent
	RentCycleEpoch         rlib.JSONTime     // Date on which rent cycle recurs. Start date for the recurring rent assessment
	UnspecifiedAdults      int64             // adults who are not accounted for in RentalAgreementPayor or RentableUser structs.  Used mostly by hotels
	UnspecifiedChildren    int64             // children who are not accounted for in RentalAgreementPayor or RentableUser structs.  Used mostly by hotels.
	Renewal                rlib.XJSONRenewal // 0 = not set, 1 = month to month automatic renewal, 2 = lease extension options
	SpecialProvisions      string            // free-form text
	LeaseType              int64             // Full Service Gross, Gross, ModifiedGross, Tripple Net
	ExpenseAdjustmentType  int64             // Base Year, No Base Year, Pass Through
	ExpensesStop           float64           // cap on the amount of oexpenses that can be passed through to the tenant
	ExpenseStopCalculation string            // note on how to determine the expense stop
	BaseYearEnd            rlib.JSONTime     // last day of the base year
	ExpenseAdjustment      rlib.JSONTime     // the next date on which an expense adjustment is due
	EstimatedCharges       float64           // a periodic fee charged to the tenant to reimburse LL for anticipated expenses
	RateChange             float64           // predetermined amount of rent increase, expressed as a percentage
	NextRateChange         rlib.JSONTime     // he next date on which a RateChange will occur
	PermittedUses          string            // indicates primary use of the space, ex: doctor's office, or warehouse/distribution, etc.
	ExclusiveUses          string            // those uses to which the tenant has the exclusive rights within a complex, ex: Trader Joe's may have the exclusive right to sell groceries
	ExtensionOption        string            // the right to extend the term of lease by giving notice to LL, ex: 2 options to extend for 5 years each
	ExtensionOptionNotice  rlib.JSONTime     // the last date by which a Tenant can give notice of their intention to exercise the right to an extension option period
	ExpansionOption        string            // the right to expand to certanin spaces that are typically contiguous to their primary space
	ExpansionOptionNotice  rlib.JSONTime     // the last date by which a Tenant can give notice of their intention to exercise the right to an Expansion Option
	RightOfFirstRefusal    string            // Tenant may have the right to purchase their premises if LL chooses to sell
	LastModTime            rlib.JSONTime     // when was this record last written
	LastModBy              int64             // employee UID (from phonebook) that modified it
	Payors                 string            // payors list attached with this RA within same time
}

// RentalAgrForm is used to save a Rental Agreement.  It holds those values
type RentalAgrForm struct {
	Recid                  int64         `json:"recid"` // this is to support the w2ui form
	RAID                   int64         // internal unique id
	RATID                  int64         // reference to Occupancy Master Agreement
	NLID                   int64         // Note ID
	AgreementStart         rlib.JSONTime // start date for rental agreement contract
	AgreementStop          rlib.JSONTime // stop date for rental agreement contract
	PossessionStart        rlib.JSONTime // start date for Occupancy
	PossessionStop         rlib.JSONTime // stop date for Occupancy
	RentStart              rlib.JSONTime // start date for Rent
	RentStop               rlib.JSONTime // stop date for Rent
	RentCycleEpoch         rlib.JSONTime // Date on which rent cycle recurs. Start date for the recurring rent assessment
	UnspecifiedAdults      int64         // adults who are not accounted for in RentalAgreementPayor or RentableUser structs.  Used mostly by hotels
	UnspecifiedChildren    int64         // children who are not accounted for in RentalAgreementPayor or RentableUser structs.  Used mostly by hotels.
	SpecialProvisions      string        // free-form text
	LeaseType              int64         // Full Service Gross, Gross, ModifiedGross, Tripple Net
	ExpenseAdjustmentType  int64         // Base Year, No Base Year, Pass Through
	ExpensesStop           float64       // cap on the amount of oexpenses that can be passed through to the tenant
	ExpenseStopCalculation string        // note on how to determine the expense stop
	BaseYearEnd            rlib.JSONTime // last day of the base year
	ExpenseAdjustment      rlib.JSONTime // the next date on which an expense adjustment is due
	EstimatedCharges       float64       // a periodic fee charged to the tenant to reimburse LL for anticipated expenses
	RateChange             float64       // predetermined amount of rent increase, expressed as a percentage
	NextRateChange         rlib.JSONTime // he next date on which a RateChange will occur
	PermittedUses          string        // indicates primary use of the space, ex: doctor's office, or warehouse/distribution, etc.
	ExclusiveUses          string        // those uses to which the tenant has the exclusive rights within a complex, ex: Trader Joe's may have the exclusive right to sell groceries
	ExtensionOption        string        // the right to extend the term of lease by giving notice to LL, ex: 2 options to extend for 5 years each
	ExtensionOptionNotice  rlib.JSONTime // the last date by which a Tenant can give notice of their intention to exercise the right to an extension option period
	ExpansionOption        string        // the right to expand to certanin spaces that are typically contiguous to their primary space
	ExpansionOptionNotice  rlib.JSONTime // the last date by which a Tenant can give notice of their intention to exercise the right to an Expansion Option
	RightOfFirstRefusal    string        // Tenant may have the right to purchase their premises if LL chooses to sell
	LastModTime            rlib.JSONTime // when was this record last written
	LastModBy              int64         // employee UID (from phonebook) that modified it
}

// RentalAgrOther is used to save a Rental Agreement.
type RentalAgrOther struct {
	BID     rlib.W2uiHTMLSelect // Business (so that we can process by Business)
	Renewal rlib.W2uiHTMLSelect // 0 = not set, 1 = month to month automatic renewal, 2 = lease extension options
}

// RentalAgrSearchResponse is the response data for a Rental Agreement Search
type RentalAgrSearchResponse struct {
	Status  string      `json:"status"`
	Total   int64       `json:"total"`
	Records []RentalAgr `json:"records"`
}

// GetRentalAgreementResponse is the response data for GetRentalAgreement
type GetRentalAgreementResponse struct {
	Status string    `json:"status"`
	Record RentalAgr `json:"record"`
}

// rentalAgrGridFieldsMap holds the map of field (to be shown on grid)
// to actual database fields, multiple db fields means combine those
var rentalAgrGridFieldsMap = map[string][]string{
	"RAID":                   {"RentalAgreement.RAID"},
	"RATID":                  {"RentalAgreement.RATID"},
	"BID":                    {"RentalAgreement.BID"},
	"NLID":                   {"RentalAgreement.NLID"},
	"AgreementStart":         {"RentalAgreement.AgreementStart"},
	"AgreementStop":          {"RentalAgreement.AgreementStop"},
	"PossessionStart":        {"RentalAgreement.PossessionStart"},
	"PossessionStop":         {"RentalAgreement.PossessionStop"},
	"RentStart":              {"RentalAgreement.RentStart"},
	"RentStop":               {"RentalAgreement.RentStop"},
	"RentCycleEpoch":         {"RentalAgreement.RentCycleEpoch"},
	"UnspecifiedAdults":      {"RentalAgreement.UnspecifiedAdults"},
	"UnspecifiedChildren":    {"RentalAgreement.UnspecifiedChildren"},
	"Renewal":                {"RentalAgreement.Renewal"},
	"SpecialProvisions":      {"RentalAgreement.SpecialProvisions"},
	"LeaseType":              {"RentalAgreement.LeaseType"},
	"ExpenseAdjustmentType":  {"RentalAgreement.ExpenseAdjustmentType"},
	"ExpensesStop":           {"RentalAgreement.ExpensesStop"},
	"ExpenseStopCalculation": {"RentalAgreement.ExpenseStopCalculation"},
	"BaseYearEnd":            {"RentalAgreement.BaseYearEnd"},
	"ExpenseAdjustment":      {"RentalAgreement.ExpenseAdjustment"},
	"EstimatedCharges":       {"RentalAgreement.EstimatedCharges"},
	"RateChange":             {"RentalAgreement.RateChange"},
	"NextRateChange":         {"RentalAgreement.NextRateChange"},
	"PermittedUses":          {"RentalAgreement.PermittedUses"},
	"ExclusiveUses":          {"RentalAgreement.ExclusiveUses"},
	"ExtensionOption":        {"RentalAgreement.ExtensionOption"},
	"ExtensionOptionNotice":  {"RentalAgreement.ExtensionOptionNotice"},
	"ExpansionOption":        {"RentalAgreement.ExpansionOption"},
	"ExpansionOptionNotice":  {"RentalAgreement.ExpansionOptionNotice"},
	"RightOfFirstRefusal":    {"RentalAgreement.RightOfFirstRefusal"},
	"LastModTime":            {"RentalAgreement.LastModTime"},
	"LastModBy":              {"RentalAgreement.LastModBy"},
	"Payors":                 {"Transactant.FirstName", "Transactant.LastName"},
}

// which fields needs to be fetched for SQL query for rental agreements
var rentalAgrQuerySelectFields = []string{
	"RentalAgreement.RAID",
	"RentalAgreement.RATID",
	// "RentalAgreement.BID",
	"RentalAgreement.NLID",
	"RentalAgreement.AgreementStart",
	"RentalAgreement.AgreementStop",
	"RentalAgreement.PossessionStart",
	"RentalAgreement.PossessionStop",
	"RentalAgreement.RentStart",
	"RentalAgreement.RentStop",
	"RentalAgreement.RentCycleEpoch",
	"RentalAgreement.UnspecifiedAdults",
	"RentalAgreement.UnspecifiedChildren",
	// "RentalAgreement.Renewal",
	"RentalAgreement.SpecialProvisions",
	"RentalAgreement.LeaseType",
	"RentalAgreement.ExpenseAdjustmentType",
	"RentalAgreement.ExpensesStop",
	"RentalAgreement.ExpenseStopCalculation",
	"RentalAgreement.BaseYearEnd",
	"RentalAgreement.ExpenseAdjustment",
	"RentalAgreement.EstimatedCharges",
	"RentalAgreement.RateChange",
	"RentalAgreement.NextRateChange",
	"RentalAgreement.PermittedUses",
	"RentalAgreement.ExclusiveUses",
	"RentalAgreement.ExtensionOption",
	"RentalAgreement.ExtensionOptionNotice",
	"RentalAgreement.ExpansionOption",
	"RentalAgreement.ExpansionOptionNotice",
	"RentalAgreement.RightOfFirstRefusal",
	"RentalAgreement.LastModTime",
	"RentalAgreement.LastModBy",
	"GROUP_CONCAT(DISTINCT CONCAT(Transactant.FirstName, ' ', Transactant.LastName) SEPARATOR ', ') AS Payors",
}

// rentalAgrRowScan scans a result from sql row and dump it in a RentalAgr struct
func rentalAgrRowScan(rows *sql.Rows, q RentalAgr) RentalAgr {
	rlib.Errcheck(rows.Scan(&q.RAID, &q.RATID, &q.NLID, &q.AgreementStart, &q.AgreementStop, &q.PossessionStart, &q.PossessionStop, &q.RentStart, &q.RentStop, &q.RentCycleEpoch, &q.UnspecifiedAdults, &q.UnspecifiedChildren /*&q.Renewal, */, &q.SpecialProvisions, &q.LeaseType, &q.ExpenseAdjustmentType, &q.ExpensesStop, &q.ExpenseStopCalculation, &q.BaseYearEnd, &q.ExpenseAdjustment, &q.EstimatedCharges, &q.RateChange, &q.NextRateChange, &q.PermittedUses, &q.ExclusiveUses, &q.ExtensionOption, &q.ExtensionOptionNotice, &q.ExpansionOption, &q.ExpansionOptionNotice, &q.RightOfFirstRefusal, &q.LastModTime, &q.LastModBy, &q.Payors))
	return q

}

// SvcSearchHandlerRentalAgr generates a report of all RentalAgreements defined business d.BID
// wsdoc {
//  @Title  Search Rental Agreements
//	@URL /v1/rentalagrs/:BUI
//  @Method  GET, POST
//	@Synopsis Return Rental Agreements that match the criteria provided.
//  @Description
//	@Input WebGridSearchRequest
//  @Response RentalAgrSearchResponse
// wsdoc }
func SvcSearchHandlerRentalAgr(w http.ResponseWriter, r *http.Request, d *ServiceData) {
	fmt.Printf("Entered SvcSearchHandlerRentalAgr\n")

	var (
		err error
		g   RentalAgrSearchResponse
		t   = time.Now()
	)

	srch := fmt.Sprintf("RentalAgreement.BID=%d AND RentalAgreement.AgreementStop>%q", d.BID, t.Format(rlib.RRDATEINPFMT)) // default WHERE clause
	order := "RentalAgreement.RAID ASC"                                                                                    // default ORDER
	// q, qw := gridBuildQuery("RentalAgreement", srch, order, d, &p)

	// get where clause and order clause for sql query
	whereClause, orderClause := GetSearchAndSortSQL(d, rentalAgrGridFieldsMap)
	if len(whereClause) > 0 {
		srch += " AND (" + whereClause + ")"
	}
	if len(orderClause) > 0 {
		order = orderClause
	}

	// Rental Agreement Query Text Template
	rentalAgrQuery := `
	SELECT
		{{.SelectClause}}
	FROM RentalAgreement
	INNER JOIN RentalAgreementPayors ON RentalAgreementPayors.RAID=RentalAgreement.RAID
	INNER JOIN Transactant ON Transactant.TCID=RentalAgreementPayors.TCID
	WHERE {{.WhereClause}}
	GROUP BY RentalAgreement.RAID
	ORDER BY {{.OrderClause}};
	`

	// will be substituted as query clauses
	qc := queryClauses{
		"SelectClause": strings.Join(rentalAgrQuerySelectFields, ","),
		"WhereClause":  srch,
		"OrderClause":  order,
	}

	// get formatted query with substitution of select, where, order clause
	q := renderSQLQuery(rentalAgrQuery, qc)
	fmt.Printf("db query = %s\n", q)

	// execute the query
	rows, err := rlib.RRdb.Dbrr.Query(q)
	rlib.Errcheck(err)
	defer rows.Close()

	i := int64(d.wsSearchReq.Offset)
	count := 0
	for rows.Next() {
		var q RentalAgr
		q.Recid = i
		q.BID = rlib.XJSONBud(fmt.Sprintf("%d", d.BID))

		// get records info in struct q
		q = rentalAgrRowScan(rows, q)

		g.Records = append(g.Records, q)
		count++ // update the count only after adding the record
		if count >= d.wsSearchReq.Limit {
			break // if we've added the max number requested, then exit
		}
		i++
	}
	// error check
	rlib.Errcheck(rows.Err())

	// get total count of results
	g.Total, err = GetQueryCount(q, qc)
	if err != nil {
		fmt.Printf("Error from GetRowCount: %s\n", err.Error())
		SvcGridErrorReturn(w, err)
		return
	}
	fmt.Printf("g.Total = %d\n", g.Total)

	// write response
	g.Status = "success"
	w.Header().Set("Content-Type", "application/json")
	SvcWriteResponse(&g, w)
}

// SvcFormHandlerRentalAgreement formats a complete data record for a person suitable for use with the w2ui Form
// For this call, we expect the URI to contain the BID and the RAID as follows:
//       0    1          2    3
// 		/v1/RentalAgrs/BID/RAID
// The server command can be:
//      get
//      save
//      delete
//-----------------------------------------------------------------------------------
func SvcFormHandlerRentalAgreement(w http.ResponseWriter, r *http.Request, d *ServiceData) {
	fmt.Printf("Entered SvcFormHandlerRentalAgreement\n")
	var err error

	if d.RAID, err = SvcExtractIDFromURI(r.RequestURI, "RAID", 3, w); err != nil {
		return
	}

	fmt.Printf("Requester UID = %d, BID = %d,  RAID = %d\n", d.UID, d.BID, d.RAID)

	switch d.wsSearchReq.Cmd {
	case "get":
		getRentalAgreement(w, r, d)
		break
	case "save":
		saveRentalAgreement(w, r, d)
		break
	default:
		err = fmt.Errorf("Unhandled command: %s", d.wsSearchReq.Cmd)
		SvcGridErrorReturn(w, err)
		return
	}
}

// wsdoc {
//  @Title  Save Rental Agreement
//	@URL /v1/rentalagr/:BUI/:RAID
//  @Method  POST
//	@Synopsis Save (create or update) a Rental Agreement
//  @Description This service returns the single-valued attributes of a Rental Agreement.
//	@Input WebGridSearchRequest
//  @Response SvcStatusResponse
// wsdoc }
func saveRentalAgreement(w http.ResponseWriter, r *http.Request, d *ServiceData) {
	funcname := "saveRentalAgreement"
	target := `"record":`
	fmt.Printf("SvcFormHandlerRentalAgreement save\n")
	fmt.Printf("record data = %s\n", d.data)
	i := strings.Index(d.data, target)
	fmt.Printf("record is at index = %d\n", i)
	if i < 0 {
		e := fmt.Errorf("saveRentalAgreement: cannot find %s in form json", target)
		SvcGridErrorReturn(w, e)
		return
	}
	s := d.data[i+len(target):]
	s = s[:len(s)-1]
	fmt.Printf("data to unmarshal is:  %s\n", s)

	//===============================================================
	//------------------------------
	// Handle all the non-list data
	//------------------------------
	var foo RentalAgrForm

	err := json.Unmarshal([]byte(s), &foo)
	if err != nil {
		e := fmt.Errorf("Error with json.Unmarshal:  %s", err.Error())
		SvcGridErrorReturn(w, e)
		return
	}
	// migrate the variables that transfer without needing special handling...
	var a rlib.RentalAgreement
	rlib.MigrateStructVals(&foo, &a)

	//---------------------------
	//  Handle all the list data
	//---------------------------
	var bar RentalAgrOther
	err = json.Unmarshal([]byte(s), &bar)
	if err != nil {
		fmt.Printf("Data unmarshal error: %s\n", err.Error())
		e := fmt.Errorf("%s: Error with json.Unmarshal:  %s", funcname, err.Error())
		SvcGridErrorReturn(w, e)
		return
	}

	var ok bool
	a.BID, ok = rlib.RRdb.BUDlist[bar.BID.ID]
	if !ok {
		e := fmt.Errorf("Could not map BID value: %s", bar.BID.ID)
		rlib.Ulog("%s", e.Error())
		SvcGridErrorReturn(w, e)
		return
	}
	a.Renewal, ok = rlib.RenewalMap[bar.Renewal.ID]
	if !ok {
		e := fmt.Errorf("could not map %s to a Renewal value", bar.Renewal.ID)
		rlib.LogAndPrintError(funcname, e)
		SvcGridErrorReturn(w, e)
		return
	}

	//===============================================================

	fmt.Printf("Update complete:  RA = %#v\n", a)

	// Now just update the database
	err = rlib.UpdateRentalAgreement(&a)
	if err != nil {
		e := fmt.Errorf("Error updating Rental Agreement RAID = %d: %s", a.RAID, err.Error())
		SvcGridErrorReturn(w, e)
		return
	}
	SvcWriteSuccessResponse(w)
}

// https://play.golang.org/p/gfOhByMroo

// wsdoc {
//  @Title  Get Rental Agreement
//	@URL /v1/rentalagr/:BUI/:RAID
//	@Method POST or GET
//	@Synopsis Get a Rental Agreement
//  @Description This service returns the single-valued attributes of a Rental Agreement.
//  @Input WebGridSearchRequest
//  @Response GetRentalAgreementResponse
// wsdoc }
func getRentalAgreement(w http.ResponseWriter, r *http.Request, d *ServiceData) {
	var g GetRentalAgreementResponse
	a, err := rlib.GetRentalAgreement(d.RAID)
	if err != nil {
		e := fmt.Errorf("getRentalAgreement: cannot read RentalAgreement RAID = %d, err = %s", d.RAID, err.Error())
		SvcGridErrorReturn(w, e)
		return
	}
	if a.RAID > 0 {
		var gg RentalAgr
		rlib.MigrateStructVals(&a, &gg)
		g.Record = gg
	}
	g.Status = "success"
	SvcWriteResponse(&g, w)
}
