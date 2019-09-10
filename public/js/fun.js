let myChart = echarts.init(document.getElementById('stats'));
let myChart1 = echarts.init(document.getElementById('stats1'));
let devices = [];
let seriesData = {};
let CalcDistance=function(second,blockTime){
    if (blockTime<=0){
        blockTime = 120;
    } else{
        blockTime /= 1000000000;
    }
    if (second<=0){
        second = 120;
    }
    let oldsecond=second,minute=0,hour=0,day=0;
    minute = parseInt(second/60); //all minutes
    second%=60;//all seconds
    if(minute>60) { //
        hour = parseInt(minute/60);//all minute
        minute%=60;//minutes
    }
    if(hour>24){//days
        day = parseInt(hour/24);
        hour%=24;//
    }
    console.log(oldsecond , blockTime);
    let allMayBlockCount = Math.ceil(parseFloat(oldsecond) / parseFloat(blockTime),0);
    if(allMayBlockCount<=0){
        allMayBlockCount += 1;
    }
    let tips = "Your devices May need "+day+"days,"+hour+"hours,"+minute+"minutes,"+second+" seconds to produce a block!<br/>";
    tips += "Your devices produce block probability is : 1/" + allMayBlockCount;
    $('#needCalc').html(tips);
    console.log(tips)
};

let renderStats = function(d){
    let series = [];
    let timespans = ["45s before","40s before","35s before","30s before","25s before","20s before","15s before","10s before","5s before","current"];
    $.each(d.devices,function(k,v){
        v.hashrate /= 1000000;//Mh/s
        if($.inArray(v.name,devices)<=-1){
            devices.push(v.name);
            seriesData[v.name] = [v.hashrate];
        } else{
            if(seriesData[v.name].length>=10){
                seriesData[v.name].shift();
            }
            seriesData[v.name].push(v.hashrate);
        }
        if(seriesData[v.name].length<10){
            timespans = timespans.slice(10-seriesData[v.name].length,10);
        }
        series.push({
            name:v.name,
            type:'line',
            stack: 'hashrate',
            areaStyle: {},
            data:seriesData[v.name]
        });
    });
    let option = {
        title: {
            text: 'Hashrate (Mh/s)'
        },
        tooltip : {
            trigger: 'axis',
            axisPointer: {
                type: 'cross',
                label: {
                    backgroundColor: '#6a7985'
                }
            }
        },
        legend: {
            data:devices
        },
        toolbox: {
            feature: {
                saveAsImage: {}
            }
        },
        grid: {
            left: '3%',
            right: '4%',
            bottom: '3%',
            containLabel: true
        },
        xAxis : [
            {
                type : 'category',
                boundaryGap : false,
                data : timespans
            }
        ],
        yAxis : [
            {
                type : 'value'
            }
        ],
        series : series
    };
    myChart.setOption(option);

    option = {
        title : {
            text: 'Mining shares stats',
            subtext: '',
            x:'center'
        },
        tooltip : {
            trigger: 'item',
            formatter: "{a} <br/>{b} : {c} ({d}%)"
        },
        legend: {
            orient: 'vertical',
            left: 'left',
            data: ['Accept','Stale','Reject']
        },
        series : [
            {
                name: 'Stats',
                type: 'pie',
                radius : '55%',
                center: ['50%', '60%'],
                data:[
                    {value:d.config.OptionConfig.Accept, name:'Accept'},
                    {value:d.config.OptionConfig.Stale, name:'Stale'},
                    {value:d.config.OptionConfig.Reject, name:'Reject'},
                ],
                itemStyle: {
                    emphasis: {
                        shadowBlur: 10,
                        shadowOffsetX: 0,
                        shadowColor: 'rgba(0, 0, 0, 0.5)'
                    }
                }
            }
        ]
    };
    myChart1.setOption(option);
    CalcDistance(d.needSec,d.blockTime);
};
let getMinerData = function () {
    $.ajax({
        url:'/miner_data',
        data:{},
        type:'get',
        dataType:'json',
    }).done(function (d) {
        let deviceHtml = '';
        $('#miner_addr').val(d.config.SoloConfig.MinerAddr);
        $('#rpc_server').val(d.config.SoloConfig.RPCServer);
        $('#rpc_username').val(d.config.SoloConfig.RPCUser);
        $('#rpc_password').val(d.config.SoloConfig.RPCPassword);
        $('#stratum_addr').val(d.config.PoolConfig.Pool);
        $('#stratum_user').val(d.config.PoolConfig.PoolUser);
        $('#stratum_pass').val(d.config.PoolConfig.PoolPassword);
        $.each(d.devices,function(k,v){
            let checked = '';
            if (v.isValid){
                checked = 'checked="checked"';
            }
            deviceHtml += '<div class="title"><h4>#'+v.id+' '+ v.name +'：</h4></div><div class="choice">  ' ;
            deviceHtml += '<label><input type="checkbox" id=d_'+v.id+' '+checked+' style="width: 20px" /> use for mining. </label><br/>';
            deviceHtml += '<span class="all">Intensity:</span></span><label><input type="text" value="'+v.global_size+'" id=gs_'+v.id+' />  </label><br/>';
            deviceHtml += '<span class="all">LocalSize:</span><label><input type="text" value="'+v.local_size+'" id=ws_'+v.id+' />  </label><br/>';
            deviceHtml += '</div>';
        });
        $('#devices').html(deviceHtml);
        renderStats(d);
    })
};

let changeDeviceStatus = function (id) {
    let ids = '';
    let inters = "";
    let lsizes = "";
    $('#devices input:checkbox').each(function() {
        if ($(this).is(':checked') === true) {
            let id = $(this).attr('id').replace("d_","");
            let intersize = $('#gs_'+id).val();
            let lsize = $('#ws_'+id).val();
            console.log(id);
            ids += id+',';
            inters += intersize+',';
            lsizes += lsize+',';
        }
    });
    if (ids.length>0){
        ids = ids.substr(ids,ids.length-1);
        inters = inters.substr(inters,inters.length-1);
        lsizes = lsizes.substr(lsizes,lsizes.length-1);
    }
    $.ajax({
        url:'/set_devices',
        data:{
            ids:ids,
            inters:inters,
            lsizes:lsizes,
        },
        type:'post',
        dataType:'json',
    }).done(function (d) {
        alert("set success");
    })
};
let submitData = function () {
    $.ajax({
        url:'/set_params',
        data:{
            miner_addr:$('#miner_addr').val(),
            rpc_server:$('#rpc_server').val(),
            rpc_username:$('#rpc_username').val(),
            rpc_password:$('#rpc_password').val(),
            stratum_addr:$('#stratum_addr').val(),
            stratum_user:$('#stratum_user').val(),
            stratum_pass:$('#stratum_pass').val()
        },
        type:'post',
        dataType:'json',
    }).done(function (d) {
        alert("set success");
    })
};
getMinerData();
function appendLog(msg) {
    console.log(msg);
}
let conn;
let testHelloWorld = function () {
    if (!conn){
        alert("WebSockets Not Support.");
        return
    }
    alert("WebSockets Success!");
    conn.send("hello world!");
};
$(function() {
    let isConnect = false;
    let reconnectServer = function () {
        if(isConnect){
            return;
        }
        isConnect = true;
        if (window["WebSocket"]) {
            let host = window.location.host;
            console.log("host:",host);
            conn = new WebSocket("ws://"+host+"/ws");
            conn.onclose = function (evt) {
                appendLog("Connection Closed.");
                setTimeout(reconnectServer,2000);
                isConnect = false;
            };
            conn.onopen = function (evt) {
                appendLog("Connection Success.");
            };
            conn.onerror = function (evt) {
                appendLog("Connection Error.");
                setTimeout(reconnectServer,2000);
                isConnect = false;
            };
            conn.onmessage = function (evt) {
                let data = $.parseJSON( evt.data );
                renderStats(data);
            };
        } else {
            appendLog("WebSockets Not Support.")
        }
    };
    reconnectServer();
});