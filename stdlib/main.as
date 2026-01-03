import "symbols.as"
import "date.as"
import "io.as"
import "code-points.as"
import "strings.as"
import "byte arrays.as"
import "arrays.as"
import "encoding.as"
import "http.as"
import "runtime.as"
import "verdex.as"

if (#_array_length(runtime.args) > 0) {
  #_run_as_script(runtime.args[0])
} else {
  #_start_repl()
}