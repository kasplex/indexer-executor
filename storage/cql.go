
////////////////////////////////
package storage

var (
    ////////////////////////////
    cqlnInitTable = []string {
        // v2.01
        "CREATE TABLE IF NOT EXISTS sttoken(p2tick ascii, tick ascii, meta ascii, minted ascii, opmod bigint, mtsmod bigint, PRIMARY KEY((p2tick), tick)) WITH CLUSTERING ORDER BY(tick ASC);",
        "CREATE TABLE IF NOT EXISTS stbalance(address ascii, tick ascii, dec tinyint, balance ascii, locked ascii, opmod bigint, PRIMARY KEY((address), tick)) WITH CLUSTERING ORDER BY(tick ASC);",
        "CREATE TABLE IF NOT EXISTS oplist(oprange bigint, opscore bigint, txid ascii, state ascii, script ascii, tickaffc ascii, addressaffc ascii, PRIMARY KEY((oprange), opscore)) WITH CLUSTERING ORDER BY(opscore ASC);",
        "CREATE TABLE IF NOT EXISTS opdata(txid ascii, state ascii, script ascii, stbefore ascii, stafter ascii, PRIMARY KEY((txid)));",
        // v2.02
        "CREATE TABLE IF NOT EXISTS stmarket(tick ascii, taddr_utxid ascii, uaddr ascii, uamt ascii, uscript ascii, tamt ascii, opadd bigint, PRIMARY KEY((tick), taddr_utxid)) WITH CLUSTERING ORDER BY(taddr_utxid ASC);",
        // v2.03
        "CREATE TABLE IF NOT EXISTS stblacklist(tick ascii, address ascii, opadd bigint, PRIMARY KEY((tick), address)) WITH CLUSTERING ORDER BY(address ASC);",
        "ALTER TABLE sttoken ADD (mod ascii, burned ascii);",
        // ...
    }
    ////////////////////////////
    cqlnGetRuntime = "SELECT * FROM runtime WHERE key=?;"
    cqlnSetRuntime = "INSERT INTO runtime (key,value1,value2,value3) VALUES (?,?,?,?);"
    ////////////////////////////
    cqlnGetVspcData = "SELECT daascore,hash,txid FROM vspc WHERE daascore IN ({daascoreIn});"
    cqlnGetVspcData2 = "SELECT daascore,hash,reorg,txid FROM vspc WHERE daascore IN ({daascoreIn});"
    ////////////////////////////
    cqlnGetTransactionData = "SELECT txid,data FROM transaction WHERE txid IN ({txidIn});"
    ////////////////////////////
    cqlnSaveStateToken = "INSERT INTO sttoken (p2tick,tick,meta,minted,opmod,mtsmod,mod,burned) VALUES (?,?,?,?,?,?,?,?);"
    cqlnDeleteStateToken = "DELETE FROM sttoken WHERE p2tick=? AND tick=?;"
    cqlnSaveStateBalance = "INSERT INTO stbalance (address,tick,dec,balance,locked,opmod) VALUES (?,?,?,?,?,?);"
    cqlnDeleteStateBalance = "DELETE FROM stbalance WHERE address=? AND tick=?;"
    cqlnSaveStateMarket = "INSERT INTO stmarket (tick,taddr_utxid,uaddr,uamt,uscript,tamt,opadd) VALUES (?,?,?,?,?,?,?);"
    cqlnDeleteStateMarket = "DELETE FROM stmarket WHERE tick=? AND taddr_utxid=?;"
    cqlnSaveStateBlacklist = "INSERT INTO stblacklist (tick,address,opadd) VALUES (?,?,?);"
    cqlnDeleteStateBlacklist = "DELETE FROM stblacklist WHERE tick=? AND address=?;"
    ////////////////////////////
    cqlnSaveOpData = "INSERT INTO opdata (txid,state,script,stbefore,stafter) VALUES (?,?,?,?,?);"
    cqlnDeleteOpData = "DELETE FROM opdata WHERE txid=?;"
    ////////////////////////////
    cqlnSaveOpList = "INSERT INTO oplist (oprange,opscore,txid,state,script,tickaffc,addressaffc) VALUES (?,?,?,?,?,?,?);"
    cqlnDeleteOpList = "DELETE FROM oplist WHERE oprange=? AND opscore=?;"
    // ...
)
