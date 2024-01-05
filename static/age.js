let fullYears = document.getElementById('fullYears');
let decimalYears = document.getElementById('decimalYears');
const birthday = new Date("2005-01-05T00:00")

function updateTime() {
    const now = new Date()
    const yearsdelta = (now.getTime() - birthday.getTime()) / 3.154e+10
    const years = yearsdelta - yearsdelta % 1
    const decimal = yearsdelta - years

    fullYears.innerHTML = "<h2>" + years.toString() + "</h2>";
    fullYears.style.textAlign = "left"
    decimalYears.innerHTML = "<h5>" + decimal.toFixed(8).substring(1) + "</h5>";
    decimalYears.style.textAlign = "center"
}

setInterval(updateTime, 100)