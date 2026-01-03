
class createCode
{
  public code = null
  public codes = #_unicode()
  constructor(name) {
    this[name] = function (string) {
      return this.codes[name] + string + this.codes.reset
    }
  }
}

immortal spawn characters = {
  [#_symbol_for("debug")]() {
    return "yo"
  }
}

for (spawn key in #_unicode()) {
  if (
    key != "newline" &&
    key != "tab" &&
    key != "reset"
    ) {
    characters[key] = (new createCode(key))[key]
  } else {
    characters[key] = #_unicode()[key]
  }
}
