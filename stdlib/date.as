
immortal spawn longMonthNames = [
  "January",
  "February",
  "March",
  "April",
  "May",
  "June",
  "July",
  "August",
  "September",
  "October",
  "November",
  "December",
]

immortal spawn weekDays = [
  "Monday",
  "Tuesday",
  "Wednesday",
  "Thursday",
  "Friday",
  "Saturday",
  "Sunday",
]

class Date {
  constructor() {}

  function getYear() {
    return #_date().getYear
  }
  function getMonth() {
    return #_date().getMonth
  }
  function getDay() {
    return #_date().getDay
  }
  function getHours() {
    return #_date().getHour
  }
  function getMinutes() {
    return #_date().getMinute
  }
  function getSeconds() {
    return #_date().getSecond
  }
  function getMilliseconds() {
    return #_date().getMillisecond
  }

  function dateToString() {
    spawn { getYear: year, getMonth: month, getDay: day, getWeekDay: weekDay } = #_date()
    return  weekDays[weekDay - 1] + ", " + day + " " + longMonthNames[month - 1] + " " + year;
  }

  function toString() {
    spawn { getYear: year, getMonth: month, getDay: day, 
    getMinute: minute, getHour: hour, getSecond: second } = #_date()
    return year + "/" + month + "/" + day + " " + hour + ":" + minute + ":" + second
  }

  function [#_symbol_for("debug")]() {
    return this.toString()
  }
}