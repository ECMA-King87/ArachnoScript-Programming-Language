static spawn Verdex = {
  WebApp: class {
    private #server = new http.Server(":2025")
    private #ast = undefined
    private module_path = #_import_meta_path()

    constructor(html, config) {
      if (typeof html != "string") {
        throw "Verdex.WebApp: first argument must be a string"
      }
      spawn meta_path = #_script_path()
      spawn html_path = #_relative_path_to_file(
        meta_path, html)
      spawn [module_path, content] = #_verdex_html(html_path)
      this.html = content
      spawn AST = #_parse_asx_module(
        #_relative_path_to_file(html_path, module_path), true)
      this.#ast = AST
      this.#routes = #_get_asx_routes()
    }

    function run() {
      this.#server.handleFunc("/verdex.js", function (w, r) {
        http.serveFile(w, r, #_relative_path_to_file(this.module_path, "./verdex.js"))
      })
      spawn module = #_compile_asx_module(this.#ast)
      for (spawn route in this.#routes) {
        this.#server.handleFunc(route, function (w, r) {
          w.writeString(#_inject_component(this.html, module))
        })
      }
      this.#server.listenAndServe()
    }
  }
}