package rlib

import (
	"database/sql"
	"sort"
	"strconv"
	"strings"
	"time"
)

// NotRentedString is the default string used
// to for the description of an unrented Rentable.
var NotRentedString = string("Unrented")

// RentRollMainRow etc all are flag to indicate types of row
const (
	RentRollMainRow       = 1 << 0
	RentRollSubTotalRow   = 1 << 1
	RentRollBlankRow      = 1 << 2
	RentRollGrandTotalRow = 1 << 3
)

// This collection of functions implements the raw data-gathering
// needed to produce a RentRoll report or interface.  These routines
// are designed to be used as shown in the pseudo code below:
//
// func myRentrollReportInterface(BID, d1,d2, iftype) {
//
//     m, n, err = GetRentRollStaticInfoMap(BID,d1,d2)           // get basic rentable info
//     m, n, err = GetRentRollVariableInfoMap(BID,d1,d2,m,n)     // Gaps, IncomeOffsets for entire collection of Rentables
//     m, n, err = GetRentRollGenTotals(BID,d1,d2,m,n)           // build subtotals and Grand Total
//
//     if iftype = UIView {
//         BuildViewInterface(m, d1,d2)
//     } else if iftype = Report {
//         BuildReport(m, d1,d2)
//     }
// }

// RentRollStaticInfo is a struct to hold the all static data
// those are received from database per row.
//
// TBD, example/test = washing machine breaks during rental period
// then offset issue ??
type RentRollStaticInfo struct {
	Recid           int64 `json:"recid"` // for webservice
	BID             int64
	RID             NullInt64
	RentableName    NullString
	RTID            NullInt64
	SqFt            NullInt64
	RentableType    NullString
	RentCycle       NullInt64
	RentCycleREP    string // rent cycle representative string
	Status          NullInt64
	Users           NullString
	RARID           NullInt64
	RAID            NullInt64
	RAIDREP         string // RAID representative string
	AgreementStart  NullDate
	AgreementStop   NullDate
	PossessionStart NullDate
	PossessionStop  NullDate
	RentStart       NullDate
	RentStop        NullDate
	Payors          NullString
	ASMID           NullInt64
	AmountDue       NullFloat64
	PaymentsApplied NullFloat64
	Description     NullString
	RentCycleGSR    NullFloat64
	PeriodGSR       NullFloat64
	IncomeOffsets   NullFloat64
	BeginReceivable float64
	DeltaReceivable float64
	EndReceivable   float64
	BeginSecDep     float64
	DeltaSecDep     float64
	EndSecDep       float64
	FLAGS           uint64 // Bits: 0 (1) = main row, 1 (2) = subtotal, 2 (4) = blank row, 3 (8) = grand total row
}

// RentRollStaticInfoRowScan scans a result from sql row and dump it in a RentRollStaticInfo struct
func RentRollStaticInfoRowScan(rows *sql.Rows, q *RentRollStaticInfo) error {
	return rows.Scan(&q.RID, &q.RentableName, &q.RTID, &q.RentableType,
		&q.RentCycle, &q.Status, &q.Users, &q.RARID, &q.RAID,
		&q.AgreementStart, &q.AgreementStop, &q.PossessionStart, &q.PossessionStop,
		&q.RentStart, &q.RentStop, &q.Payors,
		&q.ASMID, &q.AmountDue, &q.PaymentsApplied, &q.Description)
}

// RentRollStaticInfoFieldsMap holds the map of field (alias)
// to actual database field with table reference
// It could refer multiple fields
// It would be helpful in search operation with field values within db from API
var RentRollStaticInfoFieldsMap = SelectQueryFieldMap{
	"RID":             {"Rentable_CUM_RA.RID"},
	"RentableName":    {"Rentable_CUM_RA.RentableName"},
	"RTID":            {"RentableTypes.RTID"},
	"RentableType":    {"RentableTypes.Name"},
	"RentCycle":       {"RentableTypes.RentCycle"},
	"Status":          {"RentableStatus.UseStatus"},
	"Users":           {"User.FirstName", "User.LastName", "User.CompanyName"},
	"RAID":            {"Rentable_CUM_RA.RAID"},
	"AgreementStart":  {"Rentable_CUM_RA.AgreementStart"},
	"AgreementStop":   {"Rentable_CUM_RA.AgreementStop"},
	"PossessionStart": {"Rentable_CUM_RA.PossessionStart"},
	"PossessionStop":  {"Rentable_CUM_RA.PossessionStop"},
	"RentStart":       {"Rentable_CUM_RA.RentStart"},
	"RentStop":        {"Rentable_CUM_RA.RentStop"},
	"Payors":          {"Payor.FirstName", "Payor.LastName", "Payor.CompanyName"},
	"ASMID":           {"PaymentInfo.ASMID"},
	"AmountDue":       {"PaymentInfo.AmountDue"},
	"Description":     {"PaymentInfo.Description"},
}

// RentRollStaticInfoFields holds the list of fields need to be fetched
// from database for the RentRollView Query
// Field should be refer by actual db table with (.)
var RentRollStaticInfoFields = SelectQueryFields{
	"Rentable_CUM_RA.RID",
	"Rentable_CUM_RA.RentableName",
	"RentableTypes.RTID",
	"RentableTypes.Name AS RentableType",
	"RentableTypes.RentCycle",
	"RentableStatus.UseStatus AS Status",
	"GROUP_CONCAT(DISTINCT CASE WHEN User.IsCompany > 0 THEN User.CompanyName ELSE CONCAT(User.FirstName, ' ', User.LastName) END ORDER BY User.LastName ASC, User.FirstName ASC, User.CompanyName ASC SEPARATOR ', ' ) AS Users",
	"Rentable_CUM_RA.RARID",
	"Rentable_CUM_RA.RAID",
	"Rentable_CUM_RA.AgreementStart",
	"Rentable_CUM_RA.AgreementStop",
	"Rentable_CUM_RA.PossessionStart",
	"Rentable_CUM_RA.PossessionStop",
	"Rentable_CUM_RA.RentStart",
	"Rentable_CUM_RA.RentStop",
	"GROUP_CONCAT(DISTINCT CASE WHEN Payor.IsCompany > 0 THEN Payor.CompanyName ELSE CONCAT(Payor.FirstName, ' ', Payor.LastName) END ORDER BY Payor.LastName ASC, Payor.FirstName ASC, Payor.CompanyName ASC SEPARATOR ', ') AS Payors",
	"PaymentInfo.ASMID",
	"PaymentInfo.AmountDue",
	"PaymentInfo.PaymentsApplied",
	"PaymentInfo.Description",
}

// RentRollStaticInfoQuery gives the static data for rentroll rows
//-----------------------------------------------------------------------------
var RentRollStaticInfoQuery = `
SELECT
    {{.SelectClause}}
FROM
    (
        (
        /*
         *  Collect All Rentables no matter whether they got any rental agreement
         *  or not.
         */
        SELECT
            RentalAgreement.BID,
            RentalAgreement.RAID,
            RentalAgreement.AgreementStart,
            RentalAgreement.AgreementStop,
            RentalAgreement.PossessionStart,
            RentalAgreement.PossessionStop,
            RentalAgreement.RentStart,
            RentalAgreement.RentStop,
            Rentable.RID,
            Rentable.RentableName,
            RentalAgreementRentables.RARID
        FROM Rentable
            LEFT JOIN RentalAgreementRentables ON (RentalAgreementRentables.BID = Rentable.BID
                AND RentalAgreementRentables.RID = Rentable.RID
                AND @DtStart <= RentalAgreementRentables.RARDtStop
                AND @DtStop > RentalAgreementRentables.RARDtStart)
            LEFT JOIN RentalAgreement ON (RentalAgreement.BID = RentalAgreementRentables.BID
                AND RentalAgreement.RAID = RentalAgreementRentables.RAID
                AND @DtStart <= RentalAgreement.AgreementStop
                AND @DtStop > RentalAgreement.AgreementStart)
        WHERE
            Rentable.BID = @BID
        )
        UNION
        (
        /*
         *  Collect All Rental Agreements which aren't associated with any
         *  rentables.
         */
        SELECT
            RentalAgreement.BID,
            RentalAgreement.RAID,
            RentalAgreement.AgreementStart,
            RentalAgreement.AgreementStop,
            RentalAgreement.PossessionStart,
            RentalAgreement.PossessionStop,
            RentalAgreement.RentStart,
            RentalAgreement.RentStop,
            NULL AS RID,
            NULL AS RentableName,
            RentalAgreementRentables.RARID
        FROM RentalAgreement
            LEFT JOIN RentalAgreementRentables ON (RentalAgreementRentables.BID = RentalAgreement.BID
                AND RentalAgreementRentables.RAID = RentalAgreement.RAID
                AND @DtStart <= RentalAgreementRentables.RARDtStop
                AND @DtStop > RentalAgreementRentables.RARDtStart
            )
        WHERE RentalAgreement.BID = @BID
            AND RentalAgreementRentables.RAID IS NULL
            AND @DtStart <= RentalAgreement.AgreementStop
            AND @DtStop > RentalAgreement.AgreementStart
        )
    ) AS Rentable_CUM_RA
        /*
         *  Get Payors info through RentalAgreementPayors and Transactant
         */
        LEFT JOIN RentalAgreementPayors ON (Rentable_CUM_RA.RAID = RentalAgreementPayors.RAID
            AND Rentable_CUM_RA.BID = RentalAgreementPayors.BID
            AND @DtStart <= RentalAgreementPayors.DtStop
            AND @DtStop > RentalAgreementPayors.DtStart)
        LEFT JOIN Transactant AS Payor ON (Payor.TCID = RentalAgreementPayors.TCID
            AND Payor.BID = Rentable_CUM_RA.BID)
        /*
         *  RentableTypes join to get RentableType
         */
        LEFT JOIN RentableTypeRef ON (RentableTypeRef.RID = Rentable_CUM_RA.RID
            AND RentableTypeRef.BID = Rentable_CUM_RA.BID
            AND @DtStart <= RentableTypeRef.DtStop
            AND @DtStop > RentableTypeRef.DtStart
            -- Should we consider agreement dates too for comparision?
            /*AND RentableTypeRef.DtStart >= Rentable_CUM_RA.AgreementStart
            AND RentableTypeRef.DtStop <= Rentable_CUM_RA.AgreementStop*/)
        LEFT JOIN RentableTypes ON (RentableTypes.RTID = RentableTypeRef.RTID
            AND RentableTypes.BID = RentableTypeRef.BID)
        /*
         *  RentableStatus join to get the status
         */
        LEFT JOIN RentableStatus ON (RentableStatus.RID = Rentable_CUM_RA.RID
            AND RentableStatus.BID = Rentable_CUM_RA.BID
            AND @DtStart <= RentableStatus.DtStop
            AND @DtStop > RentableStatus.DtStart
            -- Should we consider agreement dates too for comparision?
            /*AND RentableStatus.DtStart >= Rentable_CUM_RA.AgreementStart
            AND RentableStatus.DtStop <= Rentable_CUM_RA.AgreementStop*/)
        /*
         *  get Users list through RentableUsers with Transactant join
         */
        LEFT JOIN RentableUsers ON (RentableUsers.RID = Rentable_CUM_RA.RID
            AND RentableUsers.RID = Rentable_CUM_RA.RID
            AND @DtStart <= RentableUsers.DtStop
            AND @DtStop > RentableUsers.DtStart
            AND RentableUsers.DtStart >= Rentable_CUM_RA.AgreementStart
            AND RentableUsers.DtStop <= Rentable_CUM_RA.AgreementStop)
        LEFT JOIN Transactant AS User ON (RentableUsers.TCID = User.TCID
            AND User.BID = Rentable_CUM_RA.BID)
        LEFT JOIN (
            /***********************************
            Assessments UNION Receipt Collection
            - - - - - - - - - - - - - - - - - */
            SELECT
                AsmRcptCollection.AmountDue AS AmountDue,
                AsmRcptCollection.ASMID,
                AsmRcptCollection.PaymentsApplied,
                AsmRcptCollection.RCPAID,
                AsmRcptCollection.RAID,
                AsmRcptCollection.RID,
                (CASE
                    WHEN AsmRcptCollection.ASMID > 0 THEN ASMARName
                    ELSE RCPTARName
                END) AS Description
            FROM
                ((
                    /*
                    Collect All Assessments with ReceiptAllocation info
                    which fall in the given report dates.
                    */
                    SELECT
                        Assessments.Amount AS AmountDue,
                        Assessments.ASMID AS ASMID,
                        SUM(DISTINCT ReceiptAllocation.Amount) as PaymentsApplied,
                        GROUP_CONCAT(DISTINCT ReceiptAllocation.RCPAID) AS RCPAID,
                        Assessments.RAID AS RAID,
                        Assessments.RID AS RID,
                        ASMAR.Name AS ASMARName,
                        NULL AS RCPTARName
                    FROM
                        Assessments
                        LEFT JOIN ReceiptAllocation ON (ReceiptAllocation.BID=Assessments.BID
                            AND ReceiptAllocation.RAID = Assessments.RAID
                            AND ReceiptAllocation.ASMID = Assessments.ASMID
                            AND @DtStart <= ReceiptAllocation.Dt
                            AND ReceiptAllocation.Dt < @DtStop)
                        LEFT JOIN Receipt ON (Receipt.BID=ReceiptAllocation.BID
                            -- AND Receipt.RAID = ReceiptAllocation.RAID // Receipt might have not updated with RAID
                            AND Receipt.RCPTID=ReceiptAllocation.RCPTID
                            AND (Receipt.FLAGS & 4) = 0
                            AND @DtStart <= Receipt.Dt
                            AND Receipt.Dt < @DtStop)
                        LEFT JOIN AR AS ASMAR ON (ASMAR.BID = Assessments.BID
                            AND ASMAR.ARID = Assessments.ARID)
                    WHERE Assessments.BID=@BID
                        AND (Assessments.RentCycle = 0 OR (Assessments.RentCycle > 0 AND Assessments.PASMID != 0))
                        AND (Assessments.FLAGS & 4) = 0
                        AND @DtStart <= Assessments.Stop
                        AND @DtStop > Assessments.Start
                    GROUP BY Assessments.ASMID
                    ORDER BY Assessments.ASMID
                ) UNION (
                    /*
                    Collect All Receipt/ReceiptAllocation of which associated assessments
                    those don't fall in the given report dates.
                    */
                    SELECT
                        NULL AS AmountDue,
                        NULL AS ASMID,
                        ReceiptAllocation.Amount AS PaymentsApplied,
                        ReceiptAllocation.RCPAID AS RCPAID,
                        ReceiptAllocation.RAID AS RAID,
                        NULL AS RID,
                        NULL AS ASMARName,
                        RCPTAR.Name AS RCPTARName
                    FROM
                        Receipt
                        INNER JOIN ReceiptAllocation ON (Receipt.BID=ReceiptAllocation.BID
                            -- AND ReceiptAllocation.RAID = Receipt.RAID // Receipt might have not updated with RAID
                            AND Receipt.RCPTID=ReceiptAllocation.RCPTID
                            AND ReceiptAllocation.ASMID > 0)
                        LEFT JOIN Assessments ON (Assessments.BID=ReceiptAllocation.BID
                            AND Assessments.RAID = ReceiptAllocation.RAID
                            AND Assessments.ASMID=ReceiptAllocation.ASMID
                            AND (Assessments.RentCycle = 0 OR (Assessments.RentCycle > 0 AND Assessments.PASMID != 0))
                            AND (Assessments.FLAGS & 4) = 0
                            AND @DtStart <= Assessments.Stop
                            AND @DtStop > Assessments.Start)
                        LEFT JOIN AR AS RCPTAR ON (RCPTAR.BID = Receipt.BID
                            AND RCPTAR.ARID = Receipt.ARID)
                    WHERE Receipt.BID=@BID
                        AND Assessments.ASMID IS NULL
                        AND (Receipt.FLAGS & 4) = 0
                        AND @DtStart <= Receipt.Dt
                        AND Receipt.Dt < @DtStop
                    GROUP BY ReceiptAllocation.RCPAID
                    ORDER BY ReceiptAllocation.RCPAID
                )) AS AsmRcptCollection
            -- Avoid any rows in which both Assessment and Receipt parts are Null
            WHERE COALESCE(AsmRcptCollection.ASMID, AsmRcptCollection.PaymentsApplied) IS NOT NULL

            /* - - - - - - - - - - - - - - - - -
            Assessments UNION Receipt Collection
            ************************************/
            ) PaymentInfo ON (PaymentInfo.RAID = Rentable_CUM_RA.RAID
                AND (CASE WHEN PaymentInfo.RID > 0 THEN PaymentInfo.RID=Rentable_CUM_RA.RID ELSE 1 END)
            )
/* GROUP BY RID, RAID, ASMID, RCPAID (In case ASMID=0)*/
GROUP BY {{.GroupClause}}
/* ORDER BY RID (if null then it would be last otherwise), RAID, AmountDue if ASMID >0 else PaymentsApplied */
ORDER BY {{.OrderClause}};
`

/*
+------+---------------------------------------------+
| NOTE | Need to take care about search operation    |
|      | As currently, we don't have the whereClause |
|      | (not required) in the viewQuery             |
+------+---------------------------------------------+
*/

// RentRollStaticInfoQueryClause - the query clause for RentRoll View
// helpful when user wants custom sorting, searching within API
var RentRollStaticInfoQueryClause = QueryClause{
	"SelectClause": strings.Join(RentRollStaticInfoFields, ","),
	"WhereClause":  "",
	"GroupClause":  "Rentable_CUM_RA.RID , Rentable_CUM_RA.RAID , (CASE WHEN PaymentInfo.ASMID > 0 THEN PaymentInfo.ASMID ELSE PaymentInfo.RCPAID END)",
	"OrderClause":  "- Rentable_CUM_RA.RID DESC , - Rentable_CUM_RA.RAID DESC , (CASE WHEN PaymentInfo.ASMID > 0 THEN PaymentInfo.AmountDue ELSE PaymentInfo.PaymentsApplied END) DESC, PaymentInfo.ASMID, PaymentInfo.RCPAID",
}

// GetRentRollStaticInfoMap returns two maps for rentroll report.
// one is of RID -> all structs that holds
// second is of RAID -> all norentable RAs staticInfo struct
//
// INPUTS
//	BID      - the business
//  startDt  - report/view start time
//  stopDt   - report/view stop time
//
// RETURNS
//  1: a map of slices of static info structs.  map key is Rentable ID (RID) - for rentable part
//  2: a map of slices of static info structs.  map key is Rentable ID (RID) - for noRentable part
//  3: any error encountered
//-----------------------------------------------------------------------------
func GetRentRollStaticInfoMap(BID int64, startDt, stopDt time.Time,
) (map[int64][]RentRollStaticInfo, map[int64][]RentRollStaticInfo, error) {

	const funcname = "GetRentRollStaticInfoMap"
	var (
		err                     error
		xbiz                    XBusiness
		rentableStaticInfoMap   = make(map[int64][]RentRollStaticInfo)
		noRentableStaticInfoMap = make(map[int64][]RentRollStaticInfo)
		d1Str                   = startDt.Format(RRDATEFMTSQL)
		d2Str                   = stopDt.Format(RRDATEFMTSQL)
	)
	Console("Entered in %s\n", funcname)

	// initialize some structures and some required things
	InitBizInternals(BID, &xbiz)

	// get formatted query
	fmtQuery := formatRentRollStaticInfoQuery(BID, startDt, stopDt, "", "", -1, -1)

	// Now, start the database transaction
	tx, err := RRdb.Dbrr.Begin()
	if err != nil {
		return rentableStaticInfoMap, noRentableStaticInfoMap, err
	}

	// set some mysql variables through `tx`
	if _, err = tx.Exec("SET @BID:=?", BID); err != nil {
		tx.Rollback()
		return rentableStaticInfoMap, noRentableStaticInfoMap, err
	}
	if _, err = tx.Exec("SET @DtStart:=?", d1Str); err != nil {
		tx.Rollback()
		return rentableStaticInfoMap, noRentableStaticInfoMap, err
	}
	if _, err = tx.Exec("SET @DtStop:=?", d2Str); err != nil {
		tx.Rollback()
		return rentableStaticInfoMap, noRentableStaticInfoMap, err
	}

	// Execute query in current transaction for Rentable section
	rrRows, err := tx.Query(fmtQuery)
	if err != nil {
		tx.Rollback()
		return rentableStaticInfoMap, noRentableStaticInfoMap, err
	}
	defer rrRows.Close()

	// ======================
	// LOOP THROUGH ALL ROWS
	// ======================
	for rrRows.Next() {
		// just assume that it is MainRow, if later encountered that it is child row
		// then "formatRentableChildRow" function would take care of it. :)
		q := RentRollStaticInfo{BID: BID}

		// database row scan
		if err = RentRollStaticInfoRowScan(rrRows, &q); err != nil { // scan next record
			return rentableStaticInfoMap, noRentableStaticInfoMap, err
		}

		if q.RID.Int64 > 0 && q.RID.Valid { // separate rentable rows in rentable staticInfo map
			rentableStaticInfoMap[q.RID.Int64] = append(rentableStaticInfoMap[q.RID.Int64], q)
		} else {
			if q.RAID.Int64 > 0 && q.RAID.Valid { // separate non rentable rows in noRentable map
				noRentableStaticInfoMap[q.RAID.Int64] = append(noRentableStaticInfoMap[q.RAID.Int64], q)
			}
		}
	}

	if err = rrRows.Err(); err != nil {
		tx.Rollback()
		return rentableStaticInfoMap, noRentableStaticInfoMap, err
	}

	// commit the transaction
	if err = tx.Commit(); err != nil {
		tx.Rollback()
		return rentableStaticInfoMap, noRentableStaticInfoMap, err
	}

	return rentableStaticInfoMap, noRentableStaticInfoMap, nil
}

// formatRentRollStaticInfoQuery returns the formatted query
// with given limit, offset if applicable.
//-----------------------------------------------------------------------------
func formatRentRollStaticInfoQuery(BID int64, d1, d2 time.Time,
	additionalWhere, orderBy string, limit, offset int) string {

	const funcname = "formatRentRollStaticInfoQuery"
	var (
		qry   = RentRollStaticInfoQuery
		qc    = GetQueryClauseCopy(RentRollStaticInfoQueryClause)
		where = qc["WhereClause"]
		order = qc["OrderClause"]
	)
	Console("Entered in : %s\n", funcname)

	// if additional conditions are provided then append
	if len(additionalWhere) > 0 {
		where += " AND (" + additionalWhere + ")"
	}
	// override orders of query results if it is given
	if len(orderBy) > 0 {
		order = orderBy
	}

	// now feed the value in queryclause
	qc["WhereClause"] = where
	qc["OrderClause"] = order

	// if limit and offset both are present then
	// we've to add limit and offset clause
	if limit > 0 && offset >= 0 {
		// if query ends with ';' then remove it
		qry = strings.TrimSuffix(strings.TrimSpace(qry), ";")

		// now add LIMIT and OFFSET clause
		qry += ` LIMIT {{.LimitClause}} OFFSET {{.OffsetClause}};`

		// feed the values of limit and offset
		qc["LimitClause"] = strconv.Itoa(limit)
		qc["OffsetClause"] = strconv.Itoa(offset)
	}

	// get formatted query with substitution of select, where, rentablesQOrder clause
	return RenderSQLQuery(qry, qc)

	// tInit := time.Now()
	// qExec, err := RRdb.Dbrr.Query(dbQry)
	// diff := time.Since(tInit)
	// Console("\nQQQQQQuery Time diff for %s is %s\n\n", rrPart, diff.String())
	// return qExec, err
}

// GetRentRollVariableInfoMap processes static info map, produces an updated
//      map. It updates the map with vacancy information for each component
//      as necessary.
//
// INPUTS
//	BID      - the business
//  startDt  - report/view start time
//  stopDt   - report/view stop time
//  m        - map created by GetRentRollStaticInfoMap
//
// RETURNS
//	1:  An updated map of slices of RentRollStaticInfo structs.
//  2:  Any error encountered
//-----------------------------------------------------------------------------
func GetRentRollVariableInfoMap(BID int64, startDt, stopDt time.Time,
	m *map[int64][]RentRollStaticInfo, n *map[int64][]RentRollStaticInfo) error {

	const funcname = "GetRentRollVariableInfoMap"

	var (
		err  error
		xbiz XBusiness
	)
	Console("Entered in %s\n", funcname)

	InitBizInternals(BID, &xbiz)

	rentrollMapGapHandler(BID, startDt, stopDt, m)
	err = rentrollMapGSRHandler(BID, startDt, stopDt, m, &xbiz)
	if err != nil {
		return err
	}

	_ = rentrollSqftHandler(BID, m, &xbiz) // ignore errors

	return nil
}

// rentrollSqftHandler get custom attribute (Square Feet)
// for each rentable
//
// INPUTS
//  BID      - the business
//  m        - pointer to map created by GetRentRollStaticInfoMap
//  xbiz     - XBusiness for getting info about RentableType and more
//
// RETURNS
//  - list of errors
//-----------------------------------------------------------------------------
func rentrollSqftHandler(BID int64,
	m *map[int64][]RentRollStaticInfo, xbiz *XBusiness,
) []error {

	const (
		funcname         = "rentrollSqftHandler"
		customAttrRTSqft = "Square Feet" // customAttrRTSqft for rentabletypes
	)

	var (
		errList = []error{}
	)
	Console("Entered in %s\n", funcname)

	// feed sqft value in first row only
	for rid := range *m {

		// only get first row from the list
		rtid := (*m)[rid][0].RTID.Int64

		// RTID should be  > 0
		if !(rtid > 0) {
			continue
		}

		if len(xbiz.RT[rtid].CA) > 0 { // if there are custom attributes
			c, ok := xbiz.RT[rtid].CA[customAttrRTSqft] // see if Square Feet is among them
			if ok {                                     // if it is...
				sqft, err := IntFromString(c.Value, "invalid customAttrRTSqft attribute")
				(*m)[rid][0].SqFt.Scan(sqft)
				if err != nil {
					Console("%s: RID: %d, RTID: %d || Error while scanning custom attribute sqft: %s\n",
						funcname, rid, rtid, err.Error())
					errList = append(errList, err)
				}
			}
		}
	}

	return errList
}

// rentrollMapGapHandler examines the supplied map and adds entries as needed to
//     describe vacancies (periods where the rentable is unrented).
//
// INPUTS
//	BID      - the business
//  startDt  - report/view start time
//  stopDt   - report/view stop time
//  m        - pointer to map created by GetRentRollStaticInfoMap
//
// RETURNS
//  no return value
//-----------------------------------------------------------------------------
func rentrollMapGapHandler(BID int64, startDt, stopDt time.Time,
	m *map[int64][]RentRollStaticInfo) {

	const funcname = "rentrollMapGapHandler"

	for rid := range *m {

		var a = []Period{}
		//--------------------------------------
		// look at all the rows for Rentable k
		//--------------------------------------
		for i := 0; i < len((*m)[rid]); i++ {
			var p = Period{
				D1: (*m)[rid][i].PossessionStart.Time,
				D2: (*m)[rid][i].PossessionStop.Time,
			}
			a = append(a, p)
		}
		b := FindGaps(&startDt, &stopDt, a) // look for gaps
		if len(b) == 0 {                    // did we find any?
			continue // NO: move on to the next Rentable
		}
		//--------------------------------------------------------------------
		// Found some gaps, create a slice of RentRollVariableData structs,
		// and add it to the map.
		//--------------------------------------------------------------------
		for i := 0; i < len(b); i++ {
			// --------------------------------------------------------
			// Get the RentableName and Rentable Type for this Rentable
			// --------------------------------------------------------
			var rName, rType string

			r := GetRentable(rid)
			rName = r.RentableName

			// NOTE: it will list down all RentableTypes, just pick first one as of now
			rts := GetRentableTypeRefsByRange(r.RID, &startDt, &stopDt)

			var rt RentableType
			if len(rts) > 0 {
				if err := GetRentableType(rts[0].RTID, &rt); err != nil {
					Console("%s: Error while getting RentableType for RID: %d", funcname, r.RID)
				}
				rType = rt.Name
			}

			//----------------------------------------------------------------
			// If the gap start and end time match the report range start and
			// end time then the Rentable is unrented for the entire period.
			// So, we will use the existing row rather than adding a new row.
			//----------------------------------------------------------------
			if b[i].D1.Equal(startDt) && b[i].D2.Equal(stopDt) {
				(*m)[rid][0].RID.Scan(rid)
				(*m)[rid][0].RentableName.Scan(rName)
				(*m)[rid][0].RentableType.Scan(rType)
				(*m)[rid][0].PossessionStart.Scan(b[i].D1) // vacancy ranges is shown in "use" column
				(*m)[rid][0].PossessionStop.Scan(b[i].D2)
				(*m)[rid][0].Description.Scan(NotRentedString)
				continue
			}
			var g RentRollStaticInfo
			g.BID = BID
			g.RID.Scan(rid)
			g.RentableName.Scan(rName)
			g.RentableType.Scan(rType)
			g.PossessionStart.Scan(b[i].D1) // vacancy ranges is shown in "use" column
			g.PossessionStop.Scan(b[i].D2)
			g.Description.Scan(NotRentedString)
			(*m)[rid] = append((*m)[rid], g)
		}
	}
}

// rentrollMapGSRHandler examines the supplied map and adds GSR information.
//
// INPUTS
//	BID      - the business
//  startDt  - report/view start time
//  stopDt   - report/view stop time
//  m        - pointer to map created by GetRentRollStaticInfoMap
//  xbiz     - XBusiness for getting info about RentableType and more
//
// RETURNS
//  any error encountered or nil if no error occurred
//-----------------------------------------------------------------------------
func rentrollMapGSRHandler(BID int64, startDt, stopDt time.Time,
	m *map[int64][]RentRollStaticInfo, xbiz *XBusiness) error {
	for k, v := range *m { // for every component

		var gsrAmt float64
		raid := int64(-1)
		for i := 0; i < len(v); i++ {
			if raid == v[i].RAID.Int64 {
				continue
			}
			raid = v[i].RAID.Int64
			//-----------------------------------------------------------------------------
			// for GSR calculation, set date range as follows...
			// start with the range of the report
			// if PossesionStart is after the report start time, then use PossessionStart
			// if PossesionStop is befor the report stop time, then use PossessionStop
			//-----------------------------------------------------------------------------
			// d1 := startDt
			// if v[i].PossessionStart.Time.After(d1) {
			// 	d1 = v[i].PossessionStart.Time
			// }
			// d2 := stopDt
			// if v[i].PossessionStop.Time.Before(d2) {
			// 	d2 = v[i].PossessionStop.Time
			// }

			d1, d2, err := ContainDateRange(&startDt, &stopDt, &v[i].PossessionStart.Time, &v[i].PossessionStop.Time)
			if err != nil {
				return err
			}

			// Console("d1 = %s, d2 = %s\n", d1.Format(RRDATEFMTSQL), d2.Format(RRDATEFMTSQL))
			gsrAmt, _, _, err = CalculateLoadedGSR(BID, k, &d1, &d2, xbiz)
			if err != nil {
				return err
			}
			gsr := GetRentableMarketRate(xbiz, k, &d1, &d2)
			v[i].RentCycleGSR = NullFloat64{Float64: gsr, Valid: true}
			v[i].PeriodGSR = NullFloat64{Float64: gsrAmt, Valid: true}
		}
	}
	return nil
}

// GetRentRollGenTotals generates the subtotal rows and grand total rows
//      of a RentRoll datastructure.
//
// INPUTS
//	BID      - the business
//  startDt  - report/view start time
//  stopDt   - report/view stop time
//  m        - pointer to rentable StaticInfo map created by GetRentRollStaticInfoMap
//  n        - pointer to noRentable staticInfo map created by GetRentRollStaticInfoMap
//  xbiz     - XBusiness for getting info about RentableType and more
//
// RETURNS
//  - grand total row
//  - total number of rows count for rentroll report
//  - total main rows count for rentroll report
//  - any error encountered or nil if no error occurred
//-----------------------------------------------------------------------------
func GetRentRollGenTotals(BID int64, startDt, stopDt time.Time,
	m *map[int64][]RentRollStaticInfo, n *map[int64][]RentRollStaticInfo,
) (RentRollStaticInfo, int64, int64, error) {

	const funcname = "GetRentRollGenTotals"
	var (
		// err           error
		totalRows     int64
		totalMainRows int64
		grandTotalRow = RentRollStaticInfo{
			BID:             BID,
			FLAGS:           RentRollGrandTotalRow,
			PeriodGSR:       NullFloat64{Valid: true, Float64: 0},
			IncomeOffsets:   NullFloat64{Valid: true, Float64: 0},
			AmountDue:       NullFloat64{Valid: true, Float64: 0},
			PaymentsApplied: NullFloat64{Valid: true, Float64: 0},
			Description:     NullString{Valid: true, String: "Grant total"},
		}
	)
	Console("Entered in %s\n", funcname)

	// sort and then format all rows for rentableStaticInfoMap as well as noRentable map
	sortAndFormatRentRollSubRows(m) // rentable StaticInfoMap
	sortAndFormatRentRollSubRows(n) // norentable StaticInfoMap

	// -------------------------
	// Rentable static info map
	// -------------------------
	for rid := range *m {

		// collection all RAID list for this component
		raidMap := make(map[int64]int64)

		// new subtotal row initialization for this component
		cmptSubTotalRow := RentRollStaticInfo{
			BID:             BID,
			FLAGS:           RentRollSubTotalRow,
			PeriodGSR:       NullFloat64{Valid: true, Float64: 0},
			IncomeOffsets:   NullFloat64{Valid: true, Float64: 0},
			AmountDue:       NullFloat64{Valid: true, Float64: 0},
			PaymentsApplied: NullFloat64{Valid: true, Float64: 0},
			Description:     NullString{Valid: true, String: "Subtotal"},
		}

		// from each row sum-up all required values
		for _, row := range (*m)[rid] {
			cmptSubTotalRow.PeriodGSR.Float64 += row.PeriodGSR.Float64
			cmptSubTotalRow.IncomeOffsets.Float64 += row.IncomeOffsets.Float64
			cmptSubTotalRow.AmountDue.Float64 += row.AmountDue.Float64
			cmptSubTotalRow.PaymentsApplied.Float64 += row.PaymentsApplied.Float64

			// raidMap, feed each RAID -> RID pair
			if _, ok := raidMap[row.RAID.Int64]; !ok && row.RAID.Int64 > 0 {
				raidMap[row.RAID.Int64] = row.RID.Int64
			}
		}

		// get all Receivables (begin, delta, ending), SecDep(begin, delta, ending) amounts
		_ = getReceivableAndSecDep(BID, startDt, stopDt, &cmptSubTotalRow, &raidMap)

		// now, append subtotal row for the current component before blank row
		(*m)[rid] = append((*m)[rid], cmptSubTotalRow)

		// add values to grand total row
		grandTotalCalculation(&grandTotalRow, &cmptSubTotalRow)

		// now, append blank row for the current component at last
		(*m)[rid] = append((*m)[rid], RentRollStaticInfo{FLAGS: RentRollBlankRow})

		// mark first rentable row as mainRow
		(*m)[rid][0].FLAGS = RentRollMainRow

		// totalRows, totalMainRows count
		totalRows += int64(len((*m)[rid]))
		totalMainRows++
	}

	// ----------------------------
	// No Rentable static info map
	// ----------------------------
	for raid := range *n {

		// collection all RAID list for this component
		raidMap := map[int64]int64{raid: 0} // RID = 0

		// new subtotal row initialization for this component
		cmptSubTotalRow := RentRollStaticInfo{
			BID:             BID,
			FLAGS:           RentRollSubTotalRow,
			PeriodGSR:       NullFloat64{Valid: true, Float64: 0},
			IncomeOffsets:   NullFloat64{Valid: true, Float64: 0},
			AmountDue:       NullFloat64{Valid: true, Float64: 0},
			PaymentsApplied: NullFloat64{Valid: true, Float64: 0},
			Description:     NullString{Valid: true, String: "Subtotal"},
		}

		// from each row sum-up all required values, at least for AmountDue, PaymentsApplied columns
		for _, row := range (*n)[raid] {
			cmptSubTotalRow.PeriodGSR.Float64 += row.PeriodGSR.Float64
			cmptSubTotalRow.IncomeOffsets.Float64 += row.IncomeOffsets.Float64
			cmptSubTotalRow.AmountDue.Float64 += row.AmountDue.Float64
			cmptSubTotalRow.PaymentsApplied.Float64 += row.PaymentsApplied.Float64
		}

		// get all Receivables (begin, delta, ending), SecDep(begin, delta, ending) amounts
		_ = getReceivableAndSecDep(BID, startDt, stopDt, &cmptSubTotalRow, &raidMap)

		// now, append subtotal row for the current component before blank row
		(*n)[raid] = append((*n)[raid], cmptSubTotalRow)

		// add values to grand total row
		grandTotalCalculation(&grandTotalRow, &cmptSubTotalRow)

		// now, append blank row for the current componene at last
		(*n)[raid] = append((*n)[raid], RentRollStaticInfo{FLAGS: RentRollBlankRow})

		// mark first no-rentable RA row as mainRow
		(*n)[raid][0].FLAGS = RentRollMainRow

		// totalRows, totalMainRows count
		totalRows += int64(len((*n)[raid]))
		totalMainRows++
	}

	return grandTotalRow, totalRows, totalMainRows, nil
}

// sortAndFormatRentRollSubRows sort all the rows
// then format the all subsequent rows
// for each component for the given static info map
//
// INPUTS
//  m        - pointer to static info map created by GetRentRollStaticInfoMap
//
// RETURNS
//  - nothing
//-----------------------------------------------------------------------------
func sortAndFormatRentRollSubRows(m *map[int64][]RentRollStaticInfo) {
	const funcname = "sortAndFormatRentRollSubRows"
	Console("Entered in %s\n", funcname)

	for k := range *m {
		// sort the list of all rows per rentable
		sort.Slice((*m)[k], func(i, j int) bool {
			if (*m)[k][i].PossessionStart.Time.Equal((*m)[k][j].PossessionStart.Time) {
				if (*m)[k][i].AmountDue.Float64 == (*m)[k][j].AmountDue.Float64 &&
					(*m)[k][i].ASMID.Int64 == (*m)[k][j].ASMID.Int64 {
					return (*m)[k][i].PaymentsApplied.Float64 > (*m)[k][j].PaymentsApplied.Float64 // descending order
				}
				return (*m)[k][i].AmountDue.Float64 > (*m)[k][j].AmountDue.Float64 // descending order
			}
			return (*m)[k][i].PossessionStart.Time.Before((*m)[k][j].PossessionStart.Time)
		})

		// ===========================
		//         FORMATTING
		// ===========================

		if (*m)[k][0].RAID.Valid && (*m)[k][0].RAID.Int64 > 0 {

			// Rent Cycle formatting for only first rentable Main Row
			for freqStr, freqNo := range CycleFreqMap {
				if (*m)[k][0].RentCycle.Int64 == freqNo && (*m)[k][0].RentCycle.Valid {
					(*m)[k][0].RentCycleREP = freqStr
				}
			}

			// Rental Agreement formatting for first row
			(*m)[k][0].RAIDREP = "RA-" + strconv.FormatInt((*m)[k][0].RAID.Int64, 10)
		}

		// skip the first row, and start to format all subsequent rows
		for rowIndex := 1; rowIndex < len((*m)[k]); rowIndex++ {
			// reset all rentable static info for subsequent rows
			(*m)[k][rowIndex].RentableName.Valid = false
			(*m)[k][rowIndex].RentableType.Valid = false
			(*m)[k][rowIndex].SqFt.Valid = false
			(*m)[k][rowIndex].RentCycleGSR.Valid = false

			// if RAID matches with previous one, then format all subsequent rows
			// for rental agreement parent-child fashion
			if (*m)[k][rowIndex].RAID.Int64 == (*m)[k][rowIndex-1].RAID.Int64 &&
				(*m)[k][rowIndex].RAID.Valid && (*m)[k][rowIndex-1].RAID.Valid {

				(*m)[k][rowIndex].PeriodGSR.Valid = false
				(*m)[k][rowIndex].RentCycle.Valid = false
				(*m)[k][rowIndex].AgreementStart.Valid = false
				(*m)[k][rowIndex].AgreementStop.Valid = false
				(*m)[k][rowIndex].PossessionStart.Valid = false
				(*m)[k][rowIndex].PossessionStop.Valid = false
				(*m)[k][rowIndex].RentStart.Valid = false
				(*m)[k][rowIndex].RentStop.Valid = false

				// it could be possible, someone introduced as payor later/removed
				if (*m)[k][rowIndex].Payors.String == (*m)[k][rowIndex-1].Payors.String {
					(*m)[k][rowIndex].Payors.Valid = false
				}

				// it could be possible, someone introduced as user later/removed
				if (*m)[k][rowIndex].Users.String == (*m)[k][rowIndex-1].Users.String {
					(*m)[k][rowIndex].Users.Valid = false
				}
			} else if (*m)[k][rowIndex].RAID.Valid {

				// Rent Cycle formatting
				for freqStr, freqNo := range CycleFreqMap {
					if (*m)[k][rowIndex].RentCycle.Int64 == freqNo && (*m)[k][rowIndex].RentCycle.Valid {
						(*m)[k][rowIndex].RentCycleREP = freqStr
					}
				}

				// Rental Agreement formatting
				(*m)[k][rowIndex].RAIDREP = "RA-" + strconv.FormatInt((*m)[k][rowIndex].RAID.Int64, 10)
			}
		}
	}
}

// grandTotalCalculation get all required field value
// from subtotal row and sum up those in grand total row
//
// INPUTS
//  grandTotalRow        - pointer to grand total row
//  subTotalRow          - pointer to sub total row
//
// RETURNS
//  - nothing
//-----------------------------------------------------------------------------
func grandTotalCalculation(grandTotalRow, subTotalRow *RentRollStaticInfo) {
	// const funcname = "grandTotalCalculation"
	// Console("Entered in %s\n", funcname)

	grandTotalRow.PeriodGSR.Float64 += subTotalRow.PeriodGSR.Float64
	grandTotalRow.IncomeOffsets.Float64 += subTotalRow.IncomeOffsets.Float64
	grandTotalRow.AmountDue.Float64 += subTotalRow.AmountDue.Float64
	grandTotalRow.PaymentsApplied.Float64 += subTotalRow.PaymentsApplied.Float64
	grandTotalRow.BeginReceivable += subTotalRow.BeginReceivable
	grandTotalRow.DeltaReceivable += subTotalRow.DeltaReceivable
	grandTotalRow.EndReceivable += subTotalRow.EndReceivable
	grandTotalRow.BeginSecDep += subTotalRow.BeginSecDep
	grandTotalRow.DeltaSecDep += subTotalRow.DeltaSecDep
	grandTotalRow.EndSecDep += subTotalRow.EndSecDep
}

// getReceivableAndSecDep calculates beginning, delta, ending receivables
// for the Receivables and Security Deposits for the given
// all Rental Agreements covered or not by with Rentables
// It will feed these amounts in the given subtotal row for the component
// of rentroll report
//
// INPUTS
//  BID      - the business
//  startDt  - report/view start time
//  stopDt   - report/view stop time
//  subTTL   - address of subtotal row for the component of rentroll report
//  raidMap  - address of map: RAID -> RID
//
// RETURNS
//  any error encountered or nil if no error occurred
//-----------------------------------------------------------------------------
func getReceivableAndSecDep(BID int64, startDt, stopDt time.Time,
	subTTL *RentRollStaticInfo, raidMap *map[int64]int64) []error {

	const funcname = "getReceivableAndSecDep"
	var (
		errList []error
		d70     = time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC)
	)
	// Console("Entered in %s.  RARBalCache size: %d\n", funcname, RARBalCacheSize())

	for raid, rid := range *raidMap {
		if !(raid > 0 && rid > 0) {
			continue // if both 0 then continue to next one
		}

		// BeginningRcv, EndingRcv
		beginningRcv, endingRcv, err := GetBeginEndRARBalance(BID, rid, raid, &startDt, &stopDt)
		if err != nil {
			Console("%s: RAID: %d, RID: %d || Error while calculating BeginningRcv, EndingRcv:: %s\n",
				funcname, raid, rid, err.Error())
			errList = append(errList, err)
		}

		/*// deltaInRcv
		deltaInRcv := (endingRcv - beginningRcv)*/

		// BeginningSecDep
		// TODO:  this should be updated to get the value on a specific date: the start date.
		//        This is only possible after we add the LedgerMarkers
		beginningSecDep, err := GetSecDepBalance(BID, raid, rid, &d70, &startDt)
		if err != nil {
			Console("%s: RAID: %d, RID: %d || Error while calculating BeginningSecDep:: %s\n",
				funcname, raid, rid, err.Error())
			errList = append(errList, err)
		}

		// Change in SecDep
		// TODO:  this should be updated to get the value on a specific date: the start date.
		//        This is only possible after we add the LedgerMarkers
		deltaInSecDep, err := GetSecDepBalance(BID, raid, rid, &startDt, &stopDt)
		if err != nil {
			Console("%s: RAID: %d, RID: %d || Error while calculating BeginningSecDep:: %s\n",
				funcname, raid, rid, err.Error())
			errList = append(errList, err)
		}

		/*// EndingSecDep
		endingSecDep := (beginningSecDep + deltaInSecDep)*/

		// now feed all those amount in subtotal row, for each iteration
		subTTL.BeginReceivable += beginningRcv
		subTTL.DeltaReceivable += (endingRcv - beginningRcv)
		subTTL.EndReceivable += endingRcv
		subTTL.BeginSecDep += beginningSecDep
		subTTL.DeltaSecDep += deltaInSecDep
		subTTL.EndSecDep += (beginningSecDep + deltaInSecDep)

		// Console("BeginSecDep = %.2f, Delta = %.2f, End = %.2f\n", subTTL.BeginSecDep, subTTL.DeltaSecDep, subTTL.EndSecDep)
	}
	// Console("Exiting %s.  RARBalCache size: %d\n", funcname, RARBalCacheSize())

	return errList
}

// GetRentRollRows returns the list of all rows for report
// It collects the rows from rentable and noRentable Maps
// with the help of three function calls
// 1. GetRentRollStaticInfoMap, 2. GetRentRollVariableInfoMap, 3. GetRentRollGenTotals
//
// INPUTS
//  BID      - the business
//  startDt  - report/view start time
//  stopDt   - report/view stop time
//  offset   - webservice offset of main rows
//  limit    - limit to send rows in the batch
//
// RETURNS
//  - list of rentroll rows
//  - total rows count
//  - total main rows count
//  - error occured during the process
//-----------------------------------------------------------------------------
func GetRentRollRows(BID int64, startDt, stopDt time.Time,
	offset, limit int) (
	[]RentRollStaticInfo, int64, int64, error) {

	const funcname = "GetRentRollRows"
	var (
		err                                error
		totalRowsCount, totalMainRowsCount int64
		rrRows                             = []RentRollStaticInfo{}
	)
	Console("Entered in %s\n", funcname)

	// pass through static field calculation
	m, n, err := GetRentRollStaticInfoMap(BID, startDt, stopDt)
	if err != nil {
		return rrRows, totalRowsCount, totalMainRowsCount, err
	}

	// go through variable field calculation
	err = GetRentRollVariableInfoMap(BID, startDt, stopDt, &m, &n)
	if err != nil {
		return rrRows, totalRowsCount, totalMainRowsCount, err
	}

	// go through total calculation
	grandTTL, totalRowsCount, totalMainRowsCount, err :=
		GetRentRollGenTotals(BID, startDt, stopDt, &m, &n)
	if err != nil {
		return rrRows, totalRowsCount, totalMainRowsCount, err
	}

	// now start to collect all rows from both map
	// to rrRows slice in sorted order of both map
	var ridList, raidList Int64Range

	for rid := range m { // sort the rentable Map
		ridList = append(ridList, rid)
	}
	sort.Sort(ridList)

	for raid := range n { // sort the non-rentable Map
		raidList = append(raidList, raid)
	}
	sort.Sort(raidList)

	if offset >= 0 && limit > 0 {
		// prepare the result array according to offset, limit values
		for mainRowsCounter := offset; mainRowsCounter < (offset + limit); mainRowsCounter++ {
			if mainRowsCounter < len(ridList) {
				rid := ridList[mainRowsCounter]
				rrRows = append(rrRows, m[rid]...)
			} else if mainRowsCounter < (len(raidList) + len(ridList)) {
				raid := raidList[mainRowsCounter-len(ridList)]
				rrRows = append(rrRows, n[raid]...)
			}
		}
	} else {
		// if no offset and limit then start to collect all rows

		// ------- iterate over rentable Map and collect all rows -----
		for _, rid := range ridList {
			rrRows = append(rrRows, m[rid]...)
		}

		// ------- iterate over non-rentable Map and collect all rows -----
		for _, raid := range raidList {
			rrRows = append(rrRows, n[raid]...)
		}
	}

	// ------- finally append grand total row -------
	// NOTE: only for first time, not for virtual scrolling
	// it might need to be changed later
	if offset <= 0 {
		rrRows = append(rrRows, grandTTL)
	}

	return rrRows, totalRowsCount, totalMainRowsCount, nil
}
