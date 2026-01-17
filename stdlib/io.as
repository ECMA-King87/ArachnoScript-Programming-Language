static spawn Console = {
  log(...data) {
    #_print(...data)
    #_print(characters.newline)
  }
}

function prompt(message, _default) {
  if ((_default ??= null, typeof _default != "string" && _default != null)) {
    throw "prompt: 2nd argument must be a string"
  }
  return #_stdin_prompt(message, _default)
}