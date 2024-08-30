
////////////////////////////////
package storage

var (
    ////////////////////////////
    cqlnInitTable = []string {
        // v2.01
        "CREATE TABLE IF NOT EXISTS sttoken(p2tick ascii, tick ascii, meta ascii, minted ascii, opmod bigint, mtsmod bigint, PRIMARY KEY((p2tick), tick)) WITH CUSTOM_PROPERTIES = {'capacity_mode':{'throughput_mode':'PAY_PER_REQUEST'}, 'point_in_time_recovery':{'status':'enabled'}, 'encryption_specification':{'encryption_type':'AWS_OWNED_KMS_KEY'}} AND CLUSTERING ORDER BY(tick ASC);",
        "CREATE TABLE IF NOT EXISTS stbalance(address ascii, tick ascii, dec tinyint, balance ascii, locked ascii, opmod bigint, PRIMARY KEY((address), tick)) WITH CUSTOM_PROPERTIES = {'capacity_mode':{'throughput_mode':'PAY_PER_REQUEST'}, 'point_in_time_recovery':{'status':'enabled'}, 'encryption_specification':{'encryption_type':'AWS_OWNED_KMS_KEY'}} AND CLUSTERING ORDER BY(tick ASC);",
        "CREATE TABLE IF NOT EXISTS oplist(oprange bigint, opscore bigint, txid ascii, state ascii, script ascii, tickaffc ascii, addressaffc ascii, PRIMARY KEY((oprange), opscore)) WITH CUSTOM_PROPERTIES = {'capacity_mode':{'throughput_mode':'PAY_PER_REQUEST'}, 'point_in_time_recovery':{'status':'enabled'}, 'encryption_specification':{'encryption_type':'AWS_OWNED_KMS_KEY'}} AND CLUSTERING ORDER BY(opscore ASC);",
        "CREATE TABLE IF NOT EXISTS opdata(txid ascii, state ascii, script ascii, stbefore ascii, stafter ascii, PRIMARY KEY((txid))) WITH CUSTOM_PROPERTIES = {'capacity_mode':{'throughput_mode':'PAY_PER_REQUEST'}, 'point_in_time_recovery':{'status':'enabled'}, 'encryption_specification':{'encryption_type':'AWS_OWNED_KMS_KEY'}};",
        // ...
    }
    ////////////////////////////
    cqlnGetRuntime = "SELECT * FROM runtime WHERE key=?;"
    cqlnSetRuntime = "INSERT INTO runtime (key,value1,value2,value3) VALUES (?,?,?,?);"
    ////////////////////////////
    cqlnGetVspcData = "SELECT daascore,hash,txid FROM vspc WHERE daascore IN ({daascoreIn});"
    ////////////////////////////
    cqlnGetTransactionData = "SELECT txid,data FROM transaction WHERE txid IN ({txidIn});"
    ////////////////////////////
    cqlnSaveStateToken = "INSERT INTO sttoken (p2tick,tick,meta,minted,opmod,mtsmod) VALUES (?,?,?,?,?,?);"
    cqlnDeleteStateToken = "DELETE FROM sttoken WHERE p2tick=? AND tick=?;"
    cqlnSaveStateBalance = "INSERT INTO stbalance (address,tick,dec,balance,locked,opmod) VALUES (?,?,?,?,?,?);"
    cqlnDeleteStateBalance = "DELETE FROM stbalance WHERE address=? AND tick=?;"
    ////////////////////////////
    cqlnSaveOpData = "INSERT INTO opdata (txid,state,script,stbefore,stafter) VALUES (?,?,?,?,?);"
    cqlnDeleteOpData = "DELETE FROM opdata WHERE txid=?;"
    ////////////////////////////
    cqlnSaveOpList = "INSERT INTO oplist (oprange,opscore,txid,state,script,tickaffc,addressaffc) VALUES (?,?,?,?,?,?,?);"
    cqlnDeleteOpList = "DELETE FROM oplist WHERE oprange=? AND opscore=?;"
    // ...
)
