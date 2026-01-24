class Array {
  public length = 0
  constructor(...elements) {
    for (immortal spawn i in elements) {
      this[i] = elements[i]
    }
    this.length = #_array_length(elements)
  }

  function at(index) {
    if (typeof index != "number") {
      throw "Array.at: index must be a number"
    }
    if (index < 0) {
      index += this.length
    }
    return this[index]
  }

  private function setLength() {
    for (immortal spawn i in this) {
      if (i > this.length - 1) {
        this.length = i + 1
      }
    }
  }

  function push(...elements) {
    for (i = 0; i < #_array_length(elements); i++) {
      this[this.length + i] = elements[i]
    }
    setLength()
  }

  function fill(value, start, end) {
    start ??= 0
    end ??= -1

    var startType = typeof start
    var endType = typeof end
    if (startType != "number" || endType != "number") {
      throw "Array.fill: start or end parameter is not a number; ("
        + startType + ", " + endType + ")";
    }
    end < 0 ? end += this.length : null;
    immortal spawn array = structuredClone(this);
    for (i = start; i <= end; i++) {
      array[i] = value
    }
    return array
  }

  function concat(array) {
    if (typeof array != "array" || !Array.isArray(array)) {
      throw "Array.concat: argument must be an array or instance of Array"
    }
  }

  private function [Symbol.iterator]() {
    spawn i = 0
    spawn self = this
    return {
      next: () => {
        return {
          done: i >= self.length,
          value: self[i++]
        }
      }
    }
  }
  
  function [#_symbol_for("debug")](char) {
    spawn col = 1;
    spawn string = "[ "
    spawn greaterThan5 = this.length > 5
    if (greaterThan5) {
      string += characters.newline + "  "
    }
    for (i = 0; i < this.length; (i++, col++)) {
      spawn lastEl = i == this.length - 1
      string += #_inspect(this[i]) + (lastEl ? " " :", ");
      if (greaterThan5 && col == 5) {
        string += characters.newline + "  "
        col = 1
      }
      if (greaterThan5 && lastEl) {
        string += characters.newline
      }
    }
    return string + "]";
  }
}

Array.isArray = function (value) {
  if (value instanceof Array) {
    return !0
  }
  return !1
}

class Iterator {
  constructor(next, self) {
    this.next = next
  }
}