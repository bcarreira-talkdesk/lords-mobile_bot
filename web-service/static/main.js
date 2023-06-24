

function loadAndRenderPlayers() {
    $.ajax({
        url: "/api/"+window.account+"/players",
        data: {
            begin:  $("#begin").val(),
            end: $("#end").val()
        }
    }).done(function( data ) {
        var selectEl = $('#player')
        selectEl.html($('<option>').val("0").text("All"))
        data.forEach(function (i) {
            selectEl.append($('<option>').val(i.user_id).text(i.user_name));
        })
    })
}

function renderCurrentDateOnEl(dateEl, addDays = 0) {
    let current_date = new Date()
    current_date.setDate(current_date.getDate() + addDays)
    let current_day = ("0" + current_date.getDate()).slice(-2);
    let current_month = ("0" + (current_date.getMonth() + 1)).slice(-2);

    let current_date_str = current_date.getFullYear()+"-"+(current_month)+"-"+(current_day) ;
    dateEl.val(current_date_str)
}

function initHuntTotalPoints() {
    var options = {
        chart: {
          id: 'totalHuntPoints',
          type: 'bar'
        },
        series: [{
          name: 'points',
          data: []
        }],
        xaxis: {
            categories: []
        }
    }
    var chart = new ApexCharts(document.querySelector("#totalHuntPoints"), options);
    chart.render();
}

function initHuntByDay() {
    var options = {
        chart: {
          id: 'huntByDay',
          type: 'line'
        },
        series: [{
            name: 'L1',
            data: []
        },{
            name: 'L2',
            data: []
        },{
            name: 'L3',
            data: []
        },{
            name: 'L4',
            data: []
        },{
            name: 'L5',
            data: []
        }],
        xaxis: {
          categories: []
        }
    }
      
    var chart = new ApexCharts(document.querySelector("#huntByDay"), options);
    chart.render();
}

function initHuntByLevel() {
    var options = {
        chart: {
          id: 'huntByLevel',
          type: 'pie'
        },
        plotOptions: {
            pie: {
                customScale: 0.8
            }
        },
        series: [],
        labels: []
    }
      
    var chart = new ApexCharts(document.querySelector("#huntByLevel"), options);
    chart.render();
}

function updateHuntTotalPoints(data) {
    ApexCharts.exec('totalHuntPoints', 'updateOptions', {
        series: [{
            name: 'points',
            data: Object.values(data).map(function(i) { return i.total_points; })
        }],
        xaxis: {
            categories: Object.keys(data)
        }
      }, false, true);
}

function updateHuntByDay(data) {
    ApexCharts.exec('huntByDay', 'updateOptions', {
        series: [{
            name: 'L1',
            data: Object.values(data).map(function(i) { return i.hunt_l1; })
        },{
            name: 'L2',
            data: Object.values(data).map(function(i) { return i.hunt_l2; })
        },{
            name: 'L3',
            data: Object.values(data).map(function(i) { return i.hunt_l3; })
        },{
            name: 'L4',
            data: Object.values(data).map(function(i) { return i.hunt_l4; })
        },{
            name: 'L5',
            data: Object.values(data).map(function(i) { return i.hunt_l5; })
        }],
        xaxis: {
          categories: Object.keys(data)
        }
      }, false, true);
}

function updateHuntByLevel(data) {
    let huntByLevel = Object.values(data).reduce(function(a, i){
        a.L1 = a.L1 + i.hunt_l1;
        a.L2 = a.L2 + i.hunt_l2;
        a.L3 = a.L3 + i.hunt_l3;
        a.L4 = a.L4 + i.hunt_l4;
        a.L5 = a.L5 + i.hunt_l5;
        return a
    }, {'L1': 0, 'L2': 0, 'L3': 0, 'L4': 0, 'L5': 0})
    ApexCharts.exec('huntByLevel', 'updateOptions', {
        series: Object.values(huntByLevel),
        labels: Object.keys(huntByLevel)
      }, false, true);
}

$("#end").change(loadAndRenderPlayers)
$("#begin").change(loadAndRenderPlayers)
$("#filter").submit(function(e){
    e.preventDefault();
    $.ajax({
        url: "/api/"+window.account+"/stats",
        data: {
            begin:  $("#begin").val(),
            end: $("#end").val(),
            user_id: $("#player option:selected").val()
        }
    }).done(function( data ) {
        var date = data.reduce(function(a, i) {
            u = a[i.date] ?? {
                                'date': i.date,
                                'hunt_l1': 0, 'hunt_l2': 0, 'hunt_l3': 0, 'hunt_l4': 0, 'hunt_l5': 0,
                                'purchase_l1': 0, 'purchase_l2': 0, 'purchase_l3': 0, 'purchase_l4': 0, 'purchase_l5': 0,
                                'hunt_points': 0, 'purchase_points': 0, 'total_points': 0
                                }
            u.hunt_l1 = u.hunt_l1 + i.hunt_l1;
            u.hunt_l2 = u.hunt_l2 + i.hunt_l2;
            u.hunt_l3 = u.hunt_l3 + i.hunt_l3;
            u.hunt_l4 = u.hunt_l4 + i.hunt_l4;
            u.hunt_l5 = u.hunt_l5 + i.hunt_l5;
            u.purchase_l1 = u.purchase_l1 + i.purchase_l1;
            u.purchase_l2 = u.purchase_l2 + i.purchase_l2;
            u.purchase_l3 = u.purchase_l3 + i.purchase_l3;
            u.purchase_l4 = u.purchase_l4 + i.purchase_l4;
            u.purchase_l5 = u.purchase_l5 + i.purchase_l5;
            u.hunt_points = u.hunt_points + i.hunt_points;
            u.purchase_points = u.purchase_points + i.purchase_points;
            u.total_points = u.total_points + i.total_points;
            a[u.date] = u
            return a
        }, {})
        updateHuntTotalPoints(date)
        updateHuntByDay(date)
        updateHuntByLevel(date)

        var datesListEl = $("#date_list > .body")
        datesListEl.html(""); // reset
        Object.values(date)
            //.sort(function(a, b){ return a.total_points < b.total_points ? 1 : -1 })
            .forEach(function(item) {
            let daterEl = $('<div>')
                            .addClass('row item')
                            .append($('<div>').text(item.date).addClass('col-2'))
                            .append($('<div>').addClass('col-3').append(
                                $('<div>').addClass('row')
                                .append($('<div>').addClass('col-1'))
                                .append($('<div>').addClass('col-2 text-center').text(item.hunt_l1))
                                .append($('<div>').addClass('col-2 text-center').text(item.hunt_l2))
                                .append($('<div>').addClass('col-2 text-center').text(item.hunt_l3))
                                .append($('<div>').addClass('col-2 text-center').text(item.hunt_l4))
                                .append($('<div>').addClass('col-2 text-center').text(item.hunt_l5))
                                .append($('<div>').addClass('col-1'))
                            ))
                            .append($('<div>').addClass('col-3').append(
                                $('<div>').addClass('row')
                                .append($('<div>').addClass('col-1'))
                                .append($('<div>').addClass('col-2 text-center').text(item.purchase_l1))
                                .append($('<div>').addClass('col-2 text-center').text(item.purchase_l2))
                                .append($('<div>').addClass('col-2 text-center').text(item.purchase_l3))
                                .append($('<div>').addClass('col-2 text-center').text(item.purchase_l4))
                                .append($('<div>').addClass('col-2 text-center').text(item.purchase_l5))
                                .append($('<div>').addClass('col-1'))
                            ))
                            .append($('<div>').addClass('col-4').append(
                                $('<div>').addClass('row')
                                .append($('<div>').addClass('col-4 text-center').text(item.hunt_points))
                                .append($('<div>').addClass('col-4 text-center').text(item.purchase_points))
                                .append($('<div>').addClass('col-4 text-center').text(item.total_points))
                            ));
            datesListEl.append(daterEl)
        })

        var users = data.reduce(function(a, i) {
            u = a[i.user_id] ?? {
                                'user_id': i.user_id, 'user_name': i.user_name, 
                                'hunt_l1': 0, 'hunt_l2': 0, 'hunt_l3': 0, 'hunt_l4': 0, 'hunt_l5': 0,
                                'purchase_l1': 0, 'purchase_l2': 0, 'purchase_l3': 0, 'purchase_l4': 0, 'purchase_l5': 0,
                                'hunt_points': 0, 'purchase_points': 0, 'total_points': 0
                                }
            u.hunt_l1 = u.hunt_l1 + i.hunt_l1;
            u.hunt_l2 = u.hunt_l2 + i.hunt_l2;
            u.hunt_l3 = u.hunt_l3 + i.hunt_l3;
            u.hunt_l4 = u.hunt_l4 + i.hunt_l4;
            u.hunt_l5 = u.hunt_l5 + i.hunt_l5;
            u.purchase_l1 = u.purchase_l1 + i.purchase_l1;
            u.purchase_l2 = u.purchase_l2 + i.purchase_l2;
            u.purchase_l3 = u.purchase_l3 + i.purchase_l3;
            u.purchase_l4 = u.purchase_l4 + i.purchase_l4;
            u.purchase_l5 = u.purchase_l5 + i.purchase_l5;
            u.hunt_points = u.hunt_points + i.hunt_points;
            u.purchase_points = u.purchase_points + i.purchase_points;
            u.total_points = u.total_points + i.total_points;
            a[u.user_id] = u
            return a
        }, {})

        var playersEl = $("#players_list > .body")
        playersEl.html(""); // reset
        Object.values(users)
            .sort(function(a, b){ return a.total_points < b.total_points ? 1 : -1 })
            .forEach(function(item) {
            let payerEl = $('<div>')
                            .addClass('row item')
                            .append($('<div>').text(item.user_name).addClass('col-2'))
                            .append($('<div>').addClass('col-3').append(
                                $('<div>').addClass('row')
                                .append($('<div>').addClass('col-1'))
                                .append($('<div>').addClass('col-2 text-center').text(item.hunt_l1))
                                .append($('<div>').addClass('col-2 text-center').text(item.hunt_l2))
                                .append($('<div>').addClass('col-2 text-center').text(item.hunt_l3))
                                .append($('<div>').addClass('col-2 text-center').text(item.hunt_l4))
                                .append($('<div>').addClass('col-2 text-center').text(item.hunt_l5))
                                .append($('<div>').addClass('col-1'))
                            ))
                            .append($('<div>').addClass('col-3').append(
                                $('<div>').addClass('row')
                                .append($('<div>').addClass('col-1'))
                                .append($('<div>').addClass('col-2 text-center').text(item.purchase_l1))
                                .append($('<div>').addClass('col-2 text-center').text(item.purchase_l2))
                                .append($('<div>').addClass('col-2 text-center').text(item.purchase_l3))
                                .append($('<div>').addClass('col-2 text-center').text(item.purchase_l4))
                                .append($('<div>').addClass('col-2 text-center').text(item.purchase_l5))
                                .append($('<div>').addClass('col-1'))
                            ))
                            .append($('<div>').addClass('col-4').append(
                                $('<div>').addClass('row')
                                .append($('<div>').addClass('col-4 text-center').text(item.hunt_points))
                                .append($('<div>').addClass('col-4 text-center').text(item.purchase_points))
                                .append($('<div>').addClass('col-4 text-center').text(item.total_points))
                            ))
            playersEl.append(payerEl)
        })
    });
})

$(document).ready(function() {
    let params = new URLSearchParams(document.location.search);
    window.account = params.get("account")
    renderCurrentDateOnEl($("#begin"), -7)
    renderCurrentDateOnEl($("#end"), 0)
    loadAndRenderPlayers()
    $("#filter").trigger('submit')
    initHuntTotalPoints()
    initHuntByDay()
    initHuntByLevel()
});
