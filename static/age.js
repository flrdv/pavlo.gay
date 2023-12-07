let fullYears = document.getElementById('fullYears');
let decimalYears = document.getElementById('decimalYears');

function updateTime() {
    const yearsSinceBirthday = (new Date() - new Date("2005-01-05")) / 3.156e+7 / 1000
    const decimal = yearsSinceBirthday - parseInt(yearsSinceBirthday)
    fullYears.innerHTML = "<h2>" + parseInt(yearsSinceBirthday).toString() + "</h2>";
    fullYears.style.textAlign = "left"
    decimalYears.innerHTML = "<h5>." + decimal.toFixed(8).substring(2) + "</h5>";
    decimalYears.style.textAlign = "center"
}

setInterval(updateTime, 100)