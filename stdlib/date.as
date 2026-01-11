
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
    return #_date().getYear()
  }
  function getMonth() {
    return #_date().getMonth()
  }
  function getDay() {
    return #_date().getDay()
  }
  function getHours() {
    return #_date().getHour()
  }
  function getMinutes() {
    return #_date().getMinute()
  }
  function getSeconds() {
    return #_date().getSecond()
  }
  function getMilliseconds() {
    return #_date().getMillisecond()
  }

  function dateToString() {
    spawn { getYear: year, getMonth: month, getDay: day, getWeekDay: weekDay } = #_date()
    return weekDays[weekDay() - 1] + ", " + day() + " " + longMonthNames[month() - 1] + " " + year();
  }

  function toString() {
    spawn { getYear, getMonth, getDay, getMinute, getHour, getSecond } = #_date()
    spawn second = getSecond();
    spawn minute = getMinute();
    spawn hour = getHour();
    spawn day = getDay();
    spawn month = getMonth();
    spawn year = getYear();
    return (day < 10 ? "0"+day : day) + "/" + (month < 10 ? "0"+month : month) + "/" + year + " " + (hour < 10 ? "0"+hour : hour) + ":" + (minute < 10 ? "0"+minute : minute) + ":" + (second < 10 ? "0"+second : second)
  }

  function [#_symbol_for("debug")]() {
    return this.toString()
  }
}

Date.now = function () { return #_time_now() }