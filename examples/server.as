spawn server = new http.Server();

server.HandleFunc("/{id}", function (w, r) {
  w.writeString(r.pathValue("id"));
})

server.listenAndServe();