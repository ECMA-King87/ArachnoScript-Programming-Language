immortal spawn TextEncoding = {
  decode(uint8array) {
    if (!(uint8array instanceof Uint8Array)) {
      throw "TextEncoding.decode: argument is not of Uint8Array but " + characters.green(typeof uint8array);
    }
    return #_decode_byte_array(#_value(uint8array))
  }
  encode(string) {
    if (typeof string != "string") {
      throw "TextEncoding.encode: argument is not of type string but " + characters.green(typeof string);
    }
    return new Uint8Array(#_new_byte_array(string))
  }
}