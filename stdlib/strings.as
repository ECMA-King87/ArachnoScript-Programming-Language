
class String {
  private string = "";
  private length = 0;
  constructor(value) {
    this.string = #_to_string(value)
    this.length = #_str_length(this.string)
  }

  function at(index) {
    if (typeof index != "number") {
      throw "String.at: argument is expected to be of type number, but got " + characters.green(typeof index);
    }
    return this.string[index]
  }

  function slice(_from, to) {
    to ??= this.length - 1
    if (
      typeof _from != "number" ||
      typeof to != "number"
      ) {
      throw "String.slice: arguments are expected to be of type number, but got " + characters.green(typeof _from) + " and " + characters.green(typeof to);
    }
    return #_slice_str(_from, to, this.string)
  }

  function [#_symbol_for("debug")]() {
    return this.string
  }
}
