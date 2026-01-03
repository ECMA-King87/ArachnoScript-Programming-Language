function isByteArray(value) {
  return #_is_byte_array(value)
}

class Uint8Array {
  public length = 0;
  private default bytes = #_new_byte_array()
  constructor(arg) {
    if (typeof arg == "number") {
      this.length = arg
    } else if (typeof arg == "array") {
      this.bytes = #_new_byte_array(arg)
      this.length = #_byte_array_length(this.bytes)
    } else if (isByteArray(arg)) {
      this.bytes = arg
      this.length = #_byte_array_length(this.bytes)
    } else {
      throw "Uint8Array: invalid argument of type " + typeof arg;
    }
  }

  function [Symbol.iterator]() {
    spawn self = this
    spawn i = 0
    return {
      next: () => {
        spawn done = i >= self.length
        spawn value = undefined
        if (!done) {
          value = #_byte_at(self.bytes, i++)
        }
        return {
          done,
          value
        }
      }
    }
  }

  function [#_symbol_for("debug")](char) {
    spawn col = 1;
    spawn string = "Uint8Array (" + this.length + ") [ "
    spawn greaterThan5 = this.length > 5
    if (greaterThan5) {
      string += characters.newline + "  "
    }
    for (i = 0; i < this.length; (i++, col++)) {
      spawn lastEl = i == this.length - 1
      string += characters.yellow(#_byte_at(this.bytes, i)) + (lastEl ? " " :", ");
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