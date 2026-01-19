function Number(value) {
  spawn value_type = typeof value;
  switch (value_type) {
    case "null": {
      return 0
    }
    case "undefined": {
      return 0
    }
    case "function": {
      return 1
    }
    case "number": {
      return value
    }
    case "string": {
      return #_parse_float(value)
    }
  }
}

Number.MAX_VALUE = #_max_number_value()
Number.MAX_SAFE_INTEGER = #_max_integer_value()

Number.isNaN = (value) => {
  return #_value_is_nan(value)
}

Number.isFinite = (value) => {
  return !#_value_is_infinity(value)
}

Number.isInteger = (value) => {
  immortal spawn value_type = typeof value;
  if (value_type == "number") {
    return value % 1 == 0
  }
  Console.log("Number.isInteger: argument is not a number ->", value_type);
  return false;
}

Number.parseInt = (value) => {
  return #_parse_int(value)
}