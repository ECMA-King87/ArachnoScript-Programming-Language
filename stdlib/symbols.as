
function Symbol(symbol) {
  if (typeof symbol == "string") {
    return #_symbol(symbol)
  }
}

Symbol.debug = #_symbol("debug");
Symbol.iterator = #_symbol("iterator");
