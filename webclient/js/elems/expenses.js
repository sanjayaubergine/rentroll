/*global
    w2ui, app, $, w2uiDateControlString, addDateNavToToolbar, console, getCurrentBusiness, getBUDfromBID,
    popupRentalAgrPicker, rafinder, form_dirty_alert, setToForm, setDateControlsInToolbar, formRefreshCallBack,
    formRecDiffer, getFormSubmitData, w2confirm, w2utils, get2XReversalSymbolHTML, getGridReversalSymbolHTML, 
*/
"use strict";
function getExpenseInitRecord(BID, BUD){
    var y = new Date();
    return {
        recid: 0,
        EXPID: 0,
        ARID: 0,
        RID: 0,
        RAID: 0,
        BID: BID,
        BUD: BUD,
        Dt: w2uiDateControlString(y),
        Amount: 0,
        AcctRule: '',
        RName: '',
        Comment: '',
        LastModTime: y.toISOString(),
        LastModBy: 0,
        CreateTS: y.toISOString(),
        CreateBy: 0,
        FLAGS: 0,
        Mode: 0,
    };
}

function renderExpReversalIcon(record /*, index, col_index*/) {
    if (typeof record === "undefined") {
        return;
    }
    if ( (record.FLAGS & app.asmFLAGS.ASMREVERSED) !== 0 ) { // if reversed then
        return '<i class="fa fa-exclamation-triangle" title="reversed" aria-hidden="true" style="color: #FFA500;"></i>';
    }
    return '';
}

function buildExpenseElements() {
    //------------------------------------------------------------------------
    //          expenseGrid
    //------------------------------------------------------------------------
    $().w2grid({
        name: 'expenseGrid',
        url: '/v1/expense',
        multiSelect: false,
        show: {
            toolbar         : true,
            footer          : true,
            toolbarAdd      : true,   // indicates if toolbar add new button is visible
            toolbarDelete   : false,   // indicates if toolbar delete button is visible
            toolbarSave     : false,   // indicates if toolbar save button is visible
            selectColumn    : false,
            expandColumn    : false,
            toolbarEdit     : false,
            toolbarSearch   : false,
            toolbarInput    : true,
            searchAll       : false,
            toolbarReload   : true,
            toolbarColumns  : true,
        },
        columns: [
            {field: 'recid',    caption: 'recid',              size: '40px',  hidden: true,  sortable: true },
            {field: 'Reversed', caption: '',                   size: '10px',  hidden: false, sortable: true, style: 'text-align: center', render: renderExpReversalIcon},
            {field: 'EXPID',    caption: 'EXPID',              size: '60px',  hidden: false, sortable: true, style: 'text-align: right'},
            {field: 'ARID',     caption: 'ARID',                              hidden: true,  sortable: false },
            {field: 'Dt',       caption: 'Date',               size: '80px',  hidden: false, sortable: true, style: 'text-align: right'},
            {field: 'ARName',   caption: 'Account Rule',       size: '200px', hidden: false, sortable: true, style: 'text-align: left'},
            {field: 'Amount',   caption: 'Amount',             size: '100px', hidden: false, sortable: true, style: 'text-align: right', render: 'money'},
            {field: 'BID',      caption: 'BID',                size: '40px',  hidden: true,  sortable: false },
            {field: 'RAID',     caption: app.sRentalAgreement, size: '125px', hidden: false, sortable: true, style: 'text-align: right'},
            {field: 'RID',      caption: 'RID',                size: '40px',  hidden: true,  sortable: false },
            {field: 'RName',    caption: app.sRentable,        size: '150px', hidden: false, sortable: true, style: 'text-align: right'},
        ],
        onClick: function(event) {
            event.onComplete = function () {
                var yes_args = [this, event.recid],
                    no_args = [this],
                    no_callBack = function(grid) {
                        grid.select(app.last.grid_sel_recid);
                        return false;
                    },
                    yes_callBack = function(grid, recid) {
                        app.last.grid_sel_recid = parseInt(recid);

                        // keep highlighting current row in any case
                        grid.select(app.last.grid_sel_recid);

                        var rec = grid.get(recid);
                        var form = "expenseForm";
                        var myurl = '/v1/expense/' + rec.BID + '/' + rec.EXPID;
                        console.log( 'calling setToForm( '+form+', ' + myurl + ')');

                        // before setting to the form, get the list of AcctRules...
                        var x = getCurrentBusiness();
                        var Bid = x.value;
                        var Bud = getBUDfromBID(Bid);
                        $.get('/v1/uival/' + x.value + '/app.ExpenseRules' )
                        .done( function(data) {
                            if (typeof data == 'string') {  // it's weird, a successful data add gets parsed as an object, an error message does not
                                app.ExpenseRules = JSON.parse(data);
                                w2ui[form].get('ARID').options.items = app.ExpenseRules[Bud];
                                w2ui[form].refresh();

                                setToForm(form, myurl, 400, true);
                            }
                            if (data.status != 'success') {
                                w2ui.expenseGrid.message(data.message);
                            }
                        })
                        .fail( function() {
                            console.log('Error getting /v1/uival/' + x.value + '/app.ExpenseRules');
                         });
                    };

                    // warn user if form content has been changed
                    form_dirty_alert(yes_callBack, no_callBack, yes_args, no_args);
            };
        },
        onAdd: function (/*event*/) {
            var yes_args = [this],
                no_callBack = function() { return false; },
                yes_callBack = function(grid) {
                    // reset it
                    app.last.grid_sel_recid = -1;
                    grid.selectNone();

                    var x = getCurrentBusiness();
                    $.get('/v1/uival/' + x.value + '/app.ExpenseRules' )
                    .done( function(data) {
                        if (typeof data == 'string') {  // it's weird, a successful data add gets parsed as an object, an error message does not
                            app.ExpenseRules = JSON.parse(data);

                            // Insert an empty record...
                            var BID=parseInt(x.value);
                            var BUD = getBUDfromBID(BID);
                            app.ridRentablePicker.BID = BID; // needed by typedown

                            var record = getExpenseInitRecord(BID, BUD);
                            // w2ui.expenseForm.fields[5].options.url = '/v1/rentablestd/' + app.ridRentablePicker.BID;
                            w2ui.expenseForm.fields[0].options.items = app.ExpenseRules[BUD];
                            w2ui.expenseForm.record = record;
                            w2ui.expenseForm.refresh();

                            setToForm('expenseForm', '/v1/expense/' + BID + '/0', 400);
                        }
                        if (data.status != 'success') {
                            w2ui.expenseForm.message(data.message);
                        }
                    })
                    .fail( function() {
                        console.log('Error getting /v1/uival/'+x.value+'/app.ExpenseRules');
                     });
                };

                // warn user if form content has been changed
                form_dirty_alert(yes_callBack, no_callBack, yes_args);
        },
        onRequest: function(/*event*/) {
            w2ui.expenseGrid.postData = {searchDtStart: app.D1, searchDtStop: app.D2};
        },
        onRefresh: function(event) {
            event.onComplete = function() {
                if (app.active_grid == this.name) {
                    if (app.new_form_rec) {
                        this.selectNone();
                    }
                    else{
                        this.select(app.last.grid_sel_recid);
                    }
                }

                if (event.target == 'monthfwd') {  // we do these tasks after monthfwd is refreshed so we know that the 2 date controls exist
                    setDateControlsInToolbar('asms');
                    w2ui.expenseGrid.postData = {searchDtStart: app.D1, searchDtStop: app.D2};
                }
            };
        }
    });

    addDateNavToToolbar('expense');

    //---------------------------------------------------------------------------------
    //          expenseForm  -  assessment epoch - this is for recurring assessments
    //---------------------------------------------------------------------------------
    $().w2form({
        name: 'expenseForm',
        style: 'border: 0px; background-color: transparent;',
        header: app.sExpense + ' Detail',
        url: '/v1/expense',
        formURL: '/webclient/html/formexpense.html',
        fields: [
            { field: 'ARID',        type: 'list',   required: true, options: { items: app.ExpenseRules }},
            { field: 'recid',       type: 'int',    required: false },
            { field: 'EXPID',       type: 'int',    required: false },
            { field: 'BID',         type: 'int',    required: true  },
            { field: 'BUD',         type: 'list',   required: true, options: {items: app.businesses} },
            { field: 'PREXPID',     type: 'int',    required: false },
            { field: 'RName',       type: 'text',   required: false },
            { field: 'RID',         type: 'int',    required: false },
            { field: 'RAID',        type: 'int',    required: false },
            { field: 'Amount',      type: 'money',  required: true  },
            { field: 'Dt',          type: 'date',   required: true  },
            { field: 'Comment',     type: 'text',   required: false },
            { field: 'LastModTime', type: 'hidden', required: false },
            { field: 'LastModBy',   type: 'hidden', required: false },
            { field: 'CreateTS',    type: 'hidden', required: false },
            { field: 'CreateBy',    type: 'hidden', required: false },
            { field: 'FLAGS',       type: 'w2int',  required: false },
        ],
        toolbar: {
            items: [
                { id: 'btnNotes', type: 'button', icon: 'fa fa-sticky-note-o' },
                { id: 'bt3',      type: 'spacer' },
                { id: 'btnClose', type: 'button', icon: 'fa fa-times' },
            ],
            onClick: function (event) {
                switch(event.target) {
                case 'btnClose':
                    var no_callBack = function() { return false; },
                        yes_callBack = function() {
                            w2ui.toplayout.hide('right',true);
                            w2ui.expenseGrid.render();
                        };
                    form_dirty_alert(yes_callBack, no_callBack);
                    break;
                }
            },
        },
        onValidate: function (event) {
            if (this.record.ARID.id === 0) {
                event.errors.push({
                    field: this.get('ARID'),
                    error: 'The Account Rule needs to be set'
                });
            }
            if (this.record.Amount < 0.01) {
                event.errors.push({
                    field: this.get('Amount'),
                    error: 'Amount must be at least $0.01'
                });
            }
        },
        actions: {
            close: function() {
                var no_callBack = function() { return false; },
                    yes_callBack = function() {
                        w2ui.toplayout.hide('right',true);
                        w2ui.expenseGrid.render();
                    };
                form_dirty_alert(yes_callBack, no_callBack);
            },
            saveadd: function() {
                var f = this,
                    grid = w2ui.expenseGrid,
                    x = getCurrentBusiness(),
                    BID=parseInt(x.value),
                    BUD = getBUDfromBID(BID);

                // clean dirty flag of form
                app.form_is_dirty = false;
                // clear the grid select recid
                app.last.grid_sel_recid  =-1;

                // unselect the record
                grid.selectNone();

                // first save the record
                f.save({}, function (data) {
                    if (data.status == 'error') {
                        console.log('ERROR: '+ data.message);
                        return;
                    }
                    // render the grid only
                    grid.render();

                    $.get('/v1/uival/' + BID + '/app.ExpenseRules' )
                    .done( function(data) {
                        if (typeof data == 'string') {  // it's weird, a successful data add gets parsed as an object, an error message does not
                            app.ExpenseRules = JSON.parse(data);
                            app.ridRentablePicker.BID = BID; // needed by typedown

                            var record = getExpenseInitRecord(BID, BUD);

                            // f.fields[5].options.url = '/v1/rentablestd/' + app.ridRentablePicker.BID;
                            f.fields[0].options.items = app.ExpenseRules[BUD];
                            f.record = record;
                            f.header = "Edit Expense (new)"; // have to provide header here, otherwise have to call refresh method twice to get this change in form
                            f.url  = "/v1/expense/" + BID + "/0";
                            f.refresh();
                        }
                        if (data.status != 'success') {
                            f.message(data.message);
                        }
                    })
                    .fail( function() {
                        console.log('Error getting /v1/uival/'+x.value+'/app.ExpenseRules');
                     }); //get assessment UI val done

                });
            },
            save: function () {
                var x = getCurrentBusiness(),
                    grid = w2ui.expenseGrid;

                grid.selectNone();
                w2ui.expenseForm.url = '/v1/expense/' + x.value + '/' + w2ui.expenseForm.record.EXPID;

                this.save({}, function (data) {
                    if (data.status == 'error') {
                        console.log('ERROR: '+ data.message);
                        return;
                    }
                    w2ui.toplayout.hide('right',true);
                    grid.render();
                });
            },
            reverse: function() {
                var form = this;

                w2confirm(reverse_confirm_options)
                .yes(function() {
                    var tgrid = w2ui.expenseGrid;
                    var params = {cmd: 'delete', formname: form.name, ID: form.record.EXPID };
                    var dat = JSON.stringify(params);
                    // Reverse receipt request
                    $.post(form.url, dat, null, "json")
                    .done(function(data) {
                        if (data.status === "error") {
                            form.error(w2utils.lang(data.message));
                            return;
                        }
                        w2ui.toplayout.hide('right',true);
                        // reversed items should not be deleted!
                        tgrid.render();
                    })
                    .fail(function(/*data*/){
                        form.error("Reverse Expense failed.");
                        return;
                    });
                })
                .no(function() {
                    return;
                });
            },
        },
        onRefresh: function(event) {
            event.onComplete = function() {
                var f = this,
                    r = f.record,
                    header = "Edit Expense ({0})";

                formRefreshCallBack(f, "EXPID", header);

                // ===========================
                // SPECIAL CASE
                // ===========================

                if (r.EXPID === 0) { // if new record then do not worry about reversed thing
                    $("#"+f.name).find("button[name=reverse]").addClass("hidden");
                    $("#"+f.name).find("button[name=save]").removeClass("hidden");
                    $("#"+f.name).find("button[name=saveadd]").removeClass("hidden");
                    $("#"+f.name).find("button[name=close]").addClass("hidden");
                    $("#"+f.name).find("#FLAGReport").addClass("hidden");
                    $("#"+f.name).find("#ExpenseInfo").addClass("hidden");

                    // ENABLE ALL INPUTS IF ALL OF THOSE HAVE BEEN DISABLED FOR REVERSED PREVIOUSLY
                    $("#"+f.name).find('input,button').prop("disabled", false);

                    return;
                } else {
                    $("#"+f.name).find("#FLAGReport").removeClass("hidden");
                    $("#"+f.name).find("#ExpenseInfo").removeClass("hidden");
                }

                // Expense Info at the top of form in white box
                var info = '<p><i class="fa fa-refresh" style="margin-right: 5px;"></i> Repeating Expense Series Definition</p>'.format(r.EXPID);
                $("#"+f.name).find("#ExpenseInfo").html(info);

                // FLAG reports
                var flag = r.FLAGS,
                    flagHTML = "";

                // check if it is reversed or not
                if ( (flag & app.asmFLAGS.ASMREVERSED) !== 0 ) { // if reversed then
                    flagHTML += "<p style='margin-bottom: 5px;'><strong>REVERSED</strong> ("+r.Comment+")</p>";
                    // reversed indication icon
                    flagHTML += get2XReversalSymbolHTML();
                    // if reversed then do not show reverse, save button
                    $("#"+f.name).find("button[name=reverse]").addClass("hidden");
                    $("#"+f.name).find("button[name=save]").addClass("hidden");
                    $("#"+f.name).find("button[name=saveadd]").addClass("hidden");
                    // if reversed then we need to show close button at the bottom of form
                    $("#"+f.name).find("button[name=close]").removeClass("hidden");

                    // *******************************************************************
                    // IF REVERSED THEN DISABLE ALL INPUTS, BUTTONS  EXCEPT close button
                    // *******************************************************************
                    $("#"+f.name).find('input,button:not([name=close])').prop("disabled", true);

                } else {
                    // show reverse, save button, hide close button
                    $("#"+f.name).find("button[name=reverse]").removeClass("hidden");
                    $("#"+f.name).find("button[name=save]").removeClass("hidden");
                    $("#"+f.name).find("button[name=saveadd]").removeClass("hidden");
                    $("#"+f.name).find("button[name=close]").addClass("hidden");

                    // ****************************************
                    // IF not REVERSED THEN ENABLE ALL INPUTS
                    // ****************************************
                    $("#"+f.name).find('input,button').prop("disabled", false);
                }

                // finally append
                // $("#"+f.name).find("#ExpFLAGReport").html(flagHTML);
                var z = document.getElementById("ExpFLAGReport");
                if (z !== null) {
                    if (r.EXPID > 0) {
                        flagHTML += "<p style='margin-bottom: 5px;'>Last Update: {0} by {1}</p>".format(r.LastModTime, r.LastModBy);
                        flagHTML += "<p>CreateTS: {0} by {1}</p>".format(r.CreateTS, r.CreateBy);
                        z.innerHTML = flagHTML;
                    } else {
                        z.innerHTML = " ";
                    }
                }
            };
        },
        onChange: function(event) {
            event.onComplete = function() {
                var f = this,
                    r = f.record;

                // copy original record temporary and reset it back after refresh event
                var temp = $.extend(true, {}, app.active_form_original);
                // finally refresh the form
                f.refresh();
                // now reset app original form record
                app.active_form_original = $.extend(true, {}, temp);

                // formRecDiffer: 1=current record, 2=original record, 3=diff object
                var diff = formRecDiffer(r, app.active_form_original, {});
                // if diff == {} then make dirty flag as false, else true
                if ($.isPlainObject(diff) && $.isEmptyObject(diff)) {
                    app.form_is_dirty = false;
                } else {
                    app.form_is_dirty = true;
                }
            };
        },
        onResize: function(event) {
            event.onComplete = function() {
                // HACK: set the height of right panel of toplayout box div and form's box div
                // this is how w2ui set the content inside box of toplayout panel, and form's main('div.w2ui-form-box')
                var h = w2ui.toplayout.get("right").height;
                $(w2ui.toplayout.get("right").content.box).height(h);
                $(this.box).find("div.w2ui-form-box").height(h);
            };
        },
        onSubmit: function(target, data) {
            delete data.postData.record.LastModTime;
            delete data.postData.record.LastModBy;
            delete data.postData.record.CreateTS;
            delete data.postData.record.CreateBy;
            // modify form data for server request
            getFormSubmitData(data.postData.record);
            console.log(data.postData.record);
        },
    });

}

function expOpenRASelect() {
    rafinder.cb = expFormRASelect;
    popupRentalAgrPicker();
}

function expFormRASelect() {
    w2ui.expenseForm.record.RAID = w2ui.rentalAgrPicker.record.RAID;
    w2ui.expenseForm.record.RName = w2ui.rentalAgrPicker.record.RentableName.text;
    w2ui.expenseForm.record.RID = w2ui.rentalAgrPicker.record.RentableName.id;
    w2ui.expenseForm.refresh();
}
