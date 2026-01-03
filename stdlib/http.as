
class ResponseWriter {
  private default writer = null

  constructor(writer) {
    if (#_is_response_writer(writer)) {
      this.writer = writer
    } else {
      throw "ResponseWriter: argument is not a response writer"
    }
  }

  function write(u8array) {
    if (u8array instanceof Uint8Array) {
      #_write_to_response_writer(this.writer, #_value(u8array))
      return;
    }
    throw "ResponseWriter.Write: argument is expected to be a Uint8Array instance"
  }

  function writeString(string) {
    if (typeof string == "string") {
      #_write_to_response_writer(this.writer, #_value(TextEncoding.encode(string)))
      return;
    }
    throw "ResponseWriter.Write: argument is expected to be a string"
  }

  function writeHeader(statusCode) {
    if (typeof statusCode == "number") {
      #_write_response_header(this.writer, statusCode)
      return
    }
    throw "ResponseWriter.Write: argument is expected to be a number but got: " + statusCode
  }

  function header() {
    return #_http_header_object(
      #_get_response_header(this.writer)
      )
  }
}

class Request {
  private default request = null
  public url = "/"
  public method = "GET"
  constructor(request) {
    if (#_is_http_request(request)) {
      this.request = request
      this.url = #_request_url(request)
      this.method = #_request_method(request)
    } else {
      throw "Request: argument is not a http request"
    }
  }

  function pathValue(string) {
    if (typeof string == "string") {
      spawn path = #_request_path_value(this.request, string)
      return path
    }
    throw "Request.pathValue: argument is expected to be a string but got: " + statusCode
  }
}


static spawn http = {
  Server: class {
    private mux = #_new_serve_mux()
    private addr = ":3457"
    constructor(addr) {
      if (typeof addr == "string" && #_str_length(addr) > 0) {
        this.addr = addr
      }
    }

    function wrapFunc(handler) {
      return function (w, r) {
        handler(new ResponseWriter(w), new Request(r))
      }
    }

    function HandleFunc(pattern, handler) {
      #_serve_mux_handle_func(this.mux, pattern, this.wrapFunc(handler))
    }

    function listenAndServe() {
      Console.log("server running on http://localhost"+this.addr)
      #_http_listen_and_serve(this.addr, this.mux)
    }
  }
  serveFile(w, r, name) {
    #_http_serve_file(#_value(w), #_value(r), name)
  }
}
