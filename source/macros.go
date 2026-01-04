package main

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var macros *Map[string, *Macro] = NewMap[string, *Macro]()

var symbol_table = NewMap[string, *Symbol]()

func init() {
	macros.set("#_print", MK_MACRO("#_print", func(args []RuntimeVal, _ *Environment, _ Pos, _ *Interpreter) RuntimeVal {
		for i := 0; i < len(args); i++ {
			PrintRtv(args[i])
			if i < len(args)-1 {
				print(" ")
			}
		}
		println()
		return undefined
	}))
	macros.set("#_symbol", MK_MACRO("#_symbol", func(args []RuntimeVal, env *Environment, pos Pos, _ *Interpreter) RuntimeVal {
		if len(args) < 1 {
			env.throwError([]string{"\x1b[31mError\x1b[0m:", "#_symbol needs one argument of type string"})
		}
		key := args[0].noAnsi()
		sym := MK_SYMBOL(key)
		symbol_table.set(key, sym)
		return sym
	}))
	macros.set("#_symbol_for", MK_MACRO("#_symbol_for", func(args []RuntimeVal, env *Environment, pos Pos, _ *Interpreter) RuntimeVal {
		if len(args) < 1 {
			env.throwError([]string{"\x1b[31mError\x1b[0m:", "#_symbol_for needs one argument of type string"})
		}
		key := args[0].noAnsi()
		sym := symbol_table.get(key)
		if sym == nil {
			sym = MK_SYMBOL(key)
			symbol_table.set(key, sym)
		}
		return sym
	}))
	code_points := map[string]string{
		"reset":     "\x1b[0m",
		"bright":    "\x1b[1m",
		"dim":       "\x1b[2m",
		"italics":   "\x1b[3m",
		"underline": "\x1b[4m",
		"red":       "\x1b[31m",
		"green":     "\x1b[32m",
		"yellow":    "\x1b[33m",
		"blue":      "\x1b[34m",
		"magenta":   "\x1b[35m",
		"cyan":      "\x1b[36m",
		"newline":   "\r\n",
		"tab":       "\t",
	}
	macros.set("#_unicode", MK_MACRO("#_unicode", func(args []RuntimeVal, env *Environment, pos Pos, _ *Interpreter) RuntimeVal {
		props := NewMap[RuntimeVal, string]()
		for k, v := range code_points {
			ml := GenerateRadix(16)
			Memory.set(ml, MK_STRING(v))
			props.set(MK_STRING(k), ml)
		}
		return MK_OBJECT(props, nil, nil)
	}))
	macros.set("#_to_string", MK_MACRO("#_to_string", func(args []RuntimeVal, env *Environment, pos Pos, r *Interpreter) RuntimeVal {
		var value RuntimeVal = undefined
		if len(args) > 0 {
			value = args[0]
		}
		return MK_STRING(value.noAnsi())
	}))
	macros.set("#_str_length", MK_MACRO("#_str_length", func(args []RuntimeVal, env *Environment, pos Pos, r *Interpreter) RuntimeVal {
		var value RuntimeVal = undefined
		length := 0
		if len(args) > 0 {
			value = args[0]
		}
		if ValueType(value) != "string" {
			length = -1
		} else {
			length = len(value.(*StringVal).value)
		}
		return MK_NUMBER(float64(length))
	}))
	macros.set("#_slice_str", MK_MACRO("#_slice_str", func(args []RuntimeVal, env *Environment, pos Pos, r *Interpreter) RuntimeVal {
		var value RuntimeVal = undefined
		length := 0
		from := 0
		to := 0
		if len(args) < 3 {
			env.throwError([]string{"#_slice_str expects 2 arguments of type (number, number, string)", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		v1 := args[0]
		v2 := args[1]
		v3 := args[2]
		if ValueType(v1) == "number" {
			from = int(v1.(*NumberVal).value)
		}
		if ValueType(v2) == "number" {
			to = int(v2.(*NumberVal).value)
		}
		if ValueType(v3) != "string" {
			env.throwError([]string{"#_slice_str expects its 3rd argument to be of type string", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		value = v3
		str := value.(*StringVal).value
		length = len(str)
		if to > length {
			to = length - 1
		}
		if to > length {
			to += length
		}
		if from < 0 {
			from += length
		}
		if from > length {
			from = length - 1
		}
		return MK_STRING(str[from : to+1])
	}))
	macros.set("#_new_byte_array", MK_MACRO("#_new_byte_array", func(args []RuntimeVal, env *Environment, pos Pos, r *Interpreter) RuntimeVal {
		if len(args) == 0 {
			return MK_RAW([]byte{})
		}
		v1 := args[0]
		var byte_array []byte = []byte{}
		switch v := v1.(type) {
		case *RawVal[byte]:
			for i := 0; i < len(args); i++ {
				arg, ok := args[i].(*RawVal[byte])
				if !ok {
					env.throwError([]string{"#_new_byte_array expects its arguments to be of type (raw [byte]) but got", fmt.Sprintf("%T", arg.Value()), SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
				}
				byte_array = append(byte_array, arg.value)
			}
		case *NumberVal:
			for i := 0; i < len(args); i++ {
				arg, ok := args[i].(*NumberVal)
				if !ok {
					env.throwError([]string{"#_new_byte_array expects its arguments to be of type (raw [byte]) but got", fmt.Sprintf("%T", arg.Value()), SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
				}
				byte_array = append(byte_array, uint8(arg.value))
			}
		case *StringVal:
			byte_array = []byte(v.value)
		default:
			env.throwError([]string{"#_new_byte_array expects its 1st argument to be of type (array [byte array]) but got", fmt.Sprintf("%T", v.Value()), SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		return MK_RAW(byte_array)
	}))
	macros.set("#_byte", MK_MACRO("#_byte", func(args []RuntimeVal, env *Environment, pos Pos, r *Interpreter) RuntimeVal {
		if len(args) < 1 {
			env.throwError([]string{"#_byte expects 1 argument of type (number | string [character])", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		v1 := args[0]
		var _byte byte
		switch b := v1.(type) {
		case *StringVal:
			_byte = byte(b.value[0])
		case *NumberVal:
			_byte = byte(b.value)
		default:
			env.throwError([]string{"#_byte: cannot convert argument of", ValueType(v1), "to byte", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		return MK_RAW(_byte)
	}))
	macros.set("#_write_byte_array", MK_MACRO("#_write_byte_array", func(args []RuntimeVal, env *Environment, pos Pos, r *Interpreter) RuntimeVal {
		if len(args) < 3 {
			env.throwError([]string{"#_write_byte_array expects 3 arguments of type (raw [byte array], raw [byte array], number)", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		v1 := args[0]
		v2 := args[1]
		v3 := args[2]
		bytes, ok := v1.(*RawVal[[]byte])
		if !ok {
			env.throwError([]string{"#_write_byte_array expects its 1st argument to be of type (raw [byte array]) but got", fmt.Sprintf("%T", v1.Value()), SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		bytes_to_write, ok := v2.(*RawVal[[]byte])
		if !ok {
			env.throwError([]string{"#_write_byte_array expects its 2nd argument to be of type (raw [byte array]) but got", fmt.Sprintf("%T", v1.Value()), SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		position, ok := v3.(*NumberVal)
		if !ok {
			env.throwError([]string{"#_write_byte_array expects its 3rd argument to be of type (number [unsigned]) but got", fmt.Sprintf("%T", v2.Value()), SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		for i := 0; i < len(bytes.value); i++ {
			bytes_to_write.value[int(position.value)+i] = bytes.value[i]
		}
		return MK_RAW([]byte{})
	}))
	macros.set("#_push_byte", MK_MACRO("#_push_byte", func(args []RuntimeVal, env *Environment, pos Pos, r *Interpreter) RuntimeVal {
		if len(args) < 2 {
			env.throwError([]string{"#_push_byte expects 2 arguments of type (raw [byte array], raw [byte])", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		v1 := args[0]
		v2 := args[1]
		bytes, ok := v1.(*RawVal[[]byte])
		if !ok {
			env.throwError([]string{"#_push_byte expects its 1st argument to be of type (raw [byte array]) but got", fmt.Sprintf("%T", v1.Value()), SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		_byte, ok := v2.(*RawVal[byte])
		if !ok {
			env.throwError([]string{"#_push_byte expects its 2nd argument to be of type (raw [byte]) but got", fmt.Sprintf("%T", v2.Value()), SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		bytes.value = append(bytes.value, _byte.value)
		return bytes
	}))
	macros.set("#_decode_byte_array", MK_MACRO("#_decode_byte_array", func(args []RuntimeVal, env *Environment, pos Pos, r *Interpreter) RuntimeVal {
		if len(args) < 1 {
			env.throwError([]string{"#_decode_byte_array expects 1 argument of type (raw [byte array])", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		v1 := args[0]
		bytes, ok := v1.(*RawVal[[]byte])
		if !ok {
			env.throwError([]string{"#_decode_byte_array expects its 1st argument to be of type (raw [byte array]) but got", fmt.Sprintf("%T", v1.Value()), SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		return MK_STRING(string(bytes.value))
	}))
	macros.set("#_is_byte_array", MK_MACRO("#_is_byte_array", func(args []RuntimeVal, env *Environment, pos Pos, r *Interpreter) RuntimeVal {
		if len(args) < 1 {
			env.throwError([]string{"#_is_byte_array expects 1 argument of type (any)", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		v1 := args[0]
		_, ok := v1.(*RawVal[[]byte])
		if ok {
			return MK_BOOL(true)
		}
		return MK_BOOL(false)
	}))
	macros.set("#_value", MK_MACRO("#_value", func(args []RuntimeVal, env *Environment, pos Pos, r *Interpreter) RuntimeVal {
		if len(args) < 1 {
			env.throwError([]string{"#_value expects 1 argument of type (any)", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		v := args[0]
		switch v := v.(type) {
		case *Instance:
			return Memory.get(v._default)
		default:
			return v
		}
	}))
	macros.set("#_byte_array_length", MK_MACRO("#_byte_array_length", func(args []RuntimeVal, env *Environment, pos Pos, r *Interpreter) RuntimeVal {
		if len(args) < 1 {
			env.throwError([]string{"#_byte_array_length expects 1 argument of type (raw [byte array])", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		v, ok := args[0].(*RawVal[[]byte])
		if !ok {
			env.throwError([]string{"#_byte_array_length expects it's 1st argument to be of type (raw [byte array])", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		return MK_NUMBER(float64(len(v.value)))
	}))
	macros.set("#_byte_at", MK_MACRO("#_byte_at", func(args []RuntimeVal, env *Environment, pos Pos, r *Interpreter) RuntimeVal {
		if len(args) < 2 {
			env.throwError([]string{"#_byte_at expects 2 arguments of type (raw [byte array], number [unsigned])", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		bytes, ok := args[0].(*RawVal[[]byte])
		if !ok {
			env.throwError([]string{"#_byte_at expects it's 1st argument to be of type (raw [byte array])", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		index, ok := args[1].(*NumberVal)
		if !ok {
			env.throwError([]string{"#_byte_at expects it's 2nd argument to be of type (number [unsigned])", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		return MK_NUMBER(float64(bytes.value[uint(index.value)]))
	}))
	macros.set("#_runtime_arguments", MK_MACRO("#_runtime_arguments", func(_ []RuntimeVal, _ *Environment, _ Pos, _ *Interpreter) RuntimeVal {
		args := []RuntimeVal{}
		// do not use range over loop
		for i := 0; i < len(arguments); i++ {
			args = append(args, MK_STRING(arguments[i]))
		}
		return MK_ARRAY(args...)
	}))
	macros.set("#_array_length", MK_MACRO("#_array_length", func(args []RuntimeVal, env *Environment, pos Pos, _ *Interpreter) RuntimeVal {
		if len(args) < 1 {
			env.throwError([]string{"#_array_length expects 1 argument of type (array)", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		array, ok := args[0].(*ArrayVal)
		if !ok {
			env.throwError([]string{"#_array_length expects it's 1st argument to be of type (array)", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		return MK_NUMBER(float64(array.elements.length))
	}))
	macros.set("#_new_serve_mux", MK_MACRO("#_new_serve_mux", func(args []RuntimeVal, env *Environment, pos Pos, _ *Interpreter) RuntimeVal {
		return MK_RAW(http.NewServeMux())
	}))
	macros.set("#_http_serve_file", MK_MACRO("#_http_serve_file", func(args []RuntimeVal, env *Environment, pos Pos, _ *Interpreter) RuntimeVal {
		if len(args) < 3 {
			env.throwError([]string{"#_http_serve_file expects 3 arguments of type (raw [serve mux], string, raw [http handler])", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		w, ok := args[0].(*RawVal[http.ResponseWriter])
		if !ok {
			env.throwError([]string{"#_http_serve_file expects it's 1st argument to be of type (raw [http response writer])", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		r, ok := args[1].(*RawVal[*http.Request])
		if !ok {
			env.throwError([]string{"#_http_serve_file expects it's 2nd argument to be of type (raw [http request])", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		name, ok := args[2].(*StringVal)
		if !ok {
			env.throwError([]string{"#_http_serve_file expects it's 3rd argument to be of type (string)", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		http.ServeFile(w.value, r.value, name.value)
		return undefined
	}))
	macros.set("#_http_serve_dir", MK_MACRO("#_http_serve_dir", func(args []RuntimeVal, env *Environment, pos Pos, _ *Interpreter) RuntimeVal {
		if len(args) < 1 {
			env.throwError([]string{"#_http_serve_dir expects 1 argument of type (string)", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		name, ok := args[0].(*StringVal)
		if !ok {
			env.throwError([]string{"#_http_serve_dir expects it's 1st argument to be of type (string)", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		return MK_RAW(http.Dir(name.value))
	}))
	macros.set("#_serve_mux_handle", MK_MACRO("#_serve_mux_handle", func(args []RuntimeVal, env *Environment, pos Pos, _ *Interpreter) RuntimeVal {
		if len(args) < 3 {
			env.throwError([]string{"#_serve_mux_handle expects 3 arguments of type (raw [serve mux], string, raw [http handler])", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		mux, ok := args[0].(*RawVal[*http.ServeMux])
		if !ok {
			env.throwError([]string{"#_serve_mux_handle expects it's 1st argument to be of type (raw [serve mux])", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		pattern, ok := args[1].(*StringVal)
		if !ok {
			env.throwError([]string{"#_serve_mux_handle expects it's 2nd argument to be of type (string)", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		handler, ok := args[2].(*RawVal[http.Handler])
		if !ok {
			env.throwError([]string{"#_serve_mux_handle expects it's 3rd argument to be of type (raw [http handler])", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		mux.value.Handle(pattern.value, handler.value)
		return undefined
	}))
	macros.set("#_http_listen_and_serve", MK_MACRO("#_http_listen_and_serve", func(args []RuntimeVal, env *Environment, pos Pos, _ *Interpreter) RuntimeVal {
		if len(args) < 2 {
			env.throwError([]string{"#_http_listen_and_serve expects 2 arguments of type (string, raw [http handler] | null)", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		pattern, ok := args[0].(*StringVal)
		if !ok {
			env.throwError([]string{"#_http_listen_and_serve expects it's 1st argument to be of type (string)", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		var handler http.Handler
		switch v := args[1].(type) {
		case *RawVal[*http.ServeMux]:
			handler = v.value
		case *RawVal[http.Handler]:
			handler = v.value
		case *NullVal:
			handler = nil
		default:
			env.throwError([]string{"#_http_listen_and_serve expects it's 2nd argument to be of type (raw [http handler] | null)", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		server := http.Server{
			Addr:    pattern.value,
			Handler: handler,
		}
		server.ListenAndServe()
		// http.ListenAndServe(pattern.value, handler)
		return undefined
	}))
	macros.set("#_serve_mux_handle_func", MK_MACRO("#_serve_mux_handle_func", func(args []RuntimeVal, env *Environment, pos Pos, runtime *Interpreter) RuntimeVal {
		if len(args) < 3 {
			env.throwError([]string{"#_serve_mux_handle_func expects 3 arguments of type (raw [serve mux], string, function)", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		mux, ok := args[0].(*RawVal[*http.ServeMux])
		if !ok {
			env.throwError([]string{"#_serve_mux_handle_func expects it's 1st argument to be of type (raw [serve mux])", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		pattern, ok := args[1].(*StringVal)
		if !ok {
			env.throwError([]string{"#_serve_mux_handle_func expects it's 2nd argument to be of type (string)", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		handler, ok := args[2].(*FunctionVal)
		if !ok {
			env.throwError([]string{"#_serve_mux_handle_func expects it's 3rd argument to be of type (function)", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		mux.value.HandleFunc(pattern.value, func(w http.ResponseWriter, r *http.Request) {
			CallFunction(handler, env, []RuntimeVal{MK_RAW(w), MK_RAW(r)}, runtime, pos)
		})
		return undefined
	}))
	macros.set("#_is_response_writer", MK_MACRO("#_is_response_writer", func(args []RuntimeVal, env *Environment, pos Pos, _ *Interpreter) RuntimeVal {
		if len(args) < 1 {
			env.throwError([]string{"#_is_response_writer expects 1 argument of type (any)", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		_, ok := args[0].(*RawVal[http.ResponseWriter])
		return MK_BOOL(ok)
	}))
	macros.set("#_is_http_request", MK_MACRO("#_is_http_request", func(args []RuntimeVal, env *Environment, pos Pos, _ *Interpreter) RuntimeVal {
		if len(args) < 1 {
			env.throwError([]string{"#_is_http_request expects 1 argument of type (any)", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		_, ok := args[0].(*RawVal[*http.Request])
		return MK_BOOL(ok)
	}))
	macros.set("#_request_path_value", MK_MACRO("#_request_path_value", func(args []RuntimeVal, env *Environment, pos Pos, _ *Interpreter) RuntimeVal {
		if len(args) < 2 {
			env.throwError([]string{"#_request_path_value expects 2 arguments of type (raw [http request], string)", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		r, ok := args[0].(*RawVal[*http.Request])
		if !ok {
			env.throwError([]string{"#_request_path_value expects it's 1st argument to be of type (raw [http request])", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		p, ok := args[1].(*StringVal)
		if !ok {
			env.throwError([]string{"#_request_path_value expects it's 2nd argument to be of type (string)", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		return MK_STRING(r.value.PathValue(p.value))
	}))
	macros.set("#_request_url", MK_MACRO("#_request_url", func(args []RuntimeVal, env *Environment, pos Pos, _ *Interpreter) RuntimeVal {
		if len(args) < 1 {
			env.throwError([]string{"#_request_url expects 1 argument of type (raw [http request])", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		r, ok := args[0].(*RawVal[*http.Request])
		if !ok {
			env.throwError([]string{"#_request_url expects it's 1st argument to be of type (raw [http request])", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		return MK_STRING(r.value.URL.Path)
	}))
	macros.set("#_request_method", MK_MACRO("#_request_method", func(args []RuntimeVal, env *Environment, pos Pos, _ *Interpreter) RuntimeVal {
		if len(args) < 1 {
			env.throwError([]string{"#_request_method expects 1 argument of type (raw [http request])", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		r, ok := args[0].(*RawVal[*http.Request])
		if !ok {
			env.throwError([]string{"#_request_method expects it's 1st argument to be of type (raw [http request])", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		method := r.value.Method
		if len(method) == 0 {
			method = "GET"
		}
		return MK_STRING(method)
	}))
	macros.set("#_write_to_response_writer", MK_MACRO("#_write_to_response_writer", func(args []RuntimeVal, env *Environment, pos Pos, r *Interpreter) RuntimeVal {
		if len(args) < 2 {
			env.throwError([]string{"#_write_to_response_writer expects 2 arguments of type (raw [response writer], raw [byte array])", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		w, ok := args[0].(*RawVal[http.ResponseWriter])
		if !ok {
			env.throwError([]string{"#_write_to_response_writer expects it's 1st argument to be of type (raw [response writer])", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		b, ok := args[1].(*RawVal[[]byte])
		if !ok {
			env.throwError([]string{"#_write_to_response_writer expects it's 2nd argument to be of type (raw [byte array])", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		i, err := w.value.Write(b.value)
		if err != nil {
			env.throwValue(MK_STRING(err.Error()), r)
		}
		return MK_NUMBER(float64(i))
	}))
	macros.set("#_write_response_header", MK_MACRO("#_write_response_header", func(args []RuntimeVal, env *Environment, pos Pos, r *Interpreter) RuntimeVal {
		if len(args) < 2 {
			env.throwError([]string{"#_write_response_header expects 2 arguments of type (raw [response writer], number)", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		w, ok := args[0].(*RawVal[http.ResponseWriter])
		if !ok {
			env.throwError([]string{"#_write_response_header expects it's 1st argument to be of type (raw [response writer])", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		b, ok := args[1].(*NumberVal)
		if !ok {
			env.throwError([]string{"#_write_response_header expects it's 2nd argument to be of type (number)", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		w.value.WriteHeader(int(b.value))
		return undefined
	}))
	macros.set("#_get_response_header", MK_MACRO("#_get_response_header", func(args []RuntimeVal, env *Environment, pos Pos, r *Interpreter) RuntimeVal {
		if len(args) < 1 {
			env.throwError([]string{"#_get_response_header expects 1 argument of type (raw [response writer])", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		w, ok := args[0].(*RawVal[http.ResponseWriter])
		if !ok {
			env.throwError([]string{"#_get_response_header expects it's 1st argument to be of type (raw [response writer])", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		return MK_RAW(w.value.Header())
	}))
	macros.set("#_http_header_object", MK_MACRO("#_http_header_object", func(args []RuntimeVal, env *Environment, pos Pos, r *Interpreter) RuntimeVal {
		if len(args) < 1 {
			env.throwError([]string{"#_http_header_object expects 1 argument of type (raw [response writer])", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		h, ok := args[0].(*RawVal[http.Header])
		if !ok {
			env.throwError([]string{"#_http_header_object expects it's 1st argument to be of type (raw [response writer])", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		header := h.value
		return createHttpHeaderObject(header, r)
	}))
	macros.set("#_date", MK_MACRO("#_date", func(args []RuntimeVal, env *Environment, pos Pos, r *Interpreter) RuntimeVal {
		now := time.Now()
		date := map[string]int{
			"getHour":        now.Hour(),
			"getMonth":       int(now.Month()),
			"getDay":         now.Day(),
			"getMinute":      now.Minute(),
			"getSecond":      now.Second(),
			"getYear":        now.Year(),
			"getMillisecond": now.Nanosecond() / 1_000_000,
			"getWeekDay":     int(now.Weekday()),
		}
		props := NewMap[RuntimeVal, string]()
		for k, v := range date {
			ml := GenerateRadix(16)
			Memory.set(ml, MK_NUMBER(float64(v)))
			props.set(MK_STRING(k), ml)
		}
		return MK_OBJECT(props, nil, nil)
	}))
	parse_method_mem_loc := GenerateRadix(16)
	macros.set("#_new_parser", MK_MACRO("#_new_parser", func(args []RuntimeVal, env *Environment, pos Pos, r *Interpreter) RuntimeVal {
		if len(args) < 2 {
			env.throwError([]string{"#_new_parser expects 2 arguments of type (string, string)", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		path, ok := args[0].(*StringVal)
		if !ok {
			env.throwError([]string{"#_new_parser expects it's 1st argument to be of type (string)", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		sourceType, ok := args[1].(*StringVal)
		if !ok {
			env.throwError([]string{"#_new_parser expects it's 2nd argument to be of type (string)", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		parser := NewParser(path.value, sourceType.value, "")
		props := NewMap[RuntimeVal, string]()
		Memory.set(parse_method_mem_loc, MK_MACRO("parse", func(args []RuntimeVal, env *Environment, pos Pos, r *Interpreter) RuntimeVal {
			if len(args) < 1 {
				env.throwError([]string{"Parser.parse (#_new_parser().parse) expects 1 argument of type (boolean)", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
			}
			main, ok := args[0].(*BoolVal)
			if !ok {
				env.throwError([]string{"Parser.parse (#_new_parser().parse) expects it's 1st argument to be of type (boolean)", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
			}
			return MK_RAW(parser.Parse(main.value))
		}))
		props.set(MK_STRING("parse"), parse_method_mem_loc)
		return MK_OBJECT(props, nil, nil)
	}))
	asx_parser := NewASXParser()
	macros.set("#_verdex_html", MK_MACRO("#_verdex_html", func(args []RuntimeVal, env *Environment, pos Pos, r *Interpreter) RuntimeVal {
		if len(args) < 1 {
			env.throwError([]string{"#_verdex_html expects 1 argument of type (string)", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		path, ok := args[0].(*StringVal)
		if !ok {
			env.throwError([]string{"#_verdex_html expects it's 1st argument to be of type (string)", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		html := ReadTextFile(path.value)
		first_line, content, found := strings.Cut(html, "\r\n")
		exp := regexp.MustCompile(`\$\{(.)+\}\$`)
		if !found || !exp.MatchString(first_line) {
			env.throwError([]string{"#_verdex_html: path to ASX module not found in", path.value, SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		module, _ := strings.CutPrefix(first_line, "${")
		module, _ = strings.CutSuffix(module, "}$")
		_1 := MK_STRING(content)
		_0 := MK_STRING(module)
		return MK_ARRAY(_0, _1)
	}))
	macros.set("#_parse_asx_module", MK_MACRO("#_parse_asx_module", func(args []RuntimeVal, env *Environment, pos Pos, r *Interpreter) RuntimeVal {
		if len(args) < 2 {
			env.throwError([]string{"#_parse_asx_module expects 2 arguments of type (string, bool)", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		path, ok := args[0].(*StringVal)
		if !ok {
			env.throwError([]string{"#_parse_asx_module expects it's 1st argument to be of type (string)", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		main, ok := args[1].(*BoolVal)
		if !ok {
			env.throwError([]string{"#_parse_asx_module expects it's 2nd argument to be of type (bool)", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		return MK_RAW(asx_parser.Parse(path.value, main.value))
	}))
	macros.set("#_compile_asx_module", MK_MACRO("#_compile_asx_module", func(args []RuntimeVal, env *Environment, pos Pos, r *Interpreter) RuntimeVal {
		if len(args) < 1 {
			env.throwError([]string{"#_compile_asx_module expects 1 argument of type (raw [asx module])", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		mod, ok := args[0].(*RawVal[*ASXModule])
		if !ok {
			env.throwError([]string{"#_compile_asx_module expects it's 1st argument to be of type (raw [asx module])", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		return MK_STRING(asx_parser.CompileASX(mod.value))
	}))
	macros.set("#_get_asx_routes", MK_MACRO("#_get_asx_routes", func(args []RuntimeVal, env *Environment, pos Pos, r *Interpreter) RuntimeVal {
		props := NewMap[RuntimeVal, string]()
		routeTable.forEach(func(key, comp string) {
			ml := GenerateRadix(16)
			Memory.set(ml, MK_STRING(comp))
			props.set(MK_STRING(key), ml)
		})
		return MK_OBJECT(props, nil, r)
	}))
	macros.set("#_inject_component", MK_MACRO("#_inject_component", func(args []RuntimeVal, env *Environment, pos Pos, r *Interpreter) RuntimeVal {
		if len(args) < 2 {
			env.throwError([]string{"#_inject_component expects 2 arguments of type (string, string)", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		html, ok := args[0].(*StringVal)
		if !ok {
			env.throwError([]string{"#_inject_component expects it's 1st argument to be of type (string)", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		module, ok := args[1].(*StringVal)
		if !ok {
			env.throwError([]string{"#_inject_component expects it's 2nd argument to be of type (string)", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		before, rest, found := strings.Cut(html.value, "</body>")
		cont := ""
		if found {
			cont = sprintf("%s<script src=\"verdex.js\"></script><script>%s\r\n__vdx_update();</script></body>%s", before, module.value, rest)
		} else {
			env.throwError([]string{"#_inject_component: could not find head tag in html", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		return MK_STRING(cont)
	}))
	macros.set("#_import_meta_path", MK_MACRO("#_import_meta_path", func(_ []RuntimeVal, env *Environment, _ Pos, _ *Interpreter) RuntimeVal {
		return MK_STRING(env.sourcePath)
	}))
	macros.set("#_as_absolute_path", MK_MACRO("#_as_absolute_path", func(args []RuntimeVal, env *Environment, pos Pos, _ *Interpreter) RuntimeVal {
		if len(args) < 1 {
			env.throwError([]string{"#_as_absolute_path expects 1 argument of type (string)", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		path, ok := args[0].(*StringVal)
		if !ok {
			env.throwError([]string{"#_as_absolute_path expects it's 1st argument to be of type (string)", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		return MK_STRING(AbsPath(path.value))
	}))
	macros.set("#_relative_path_to_file", MK_MACRO("#_relative_path_to_file", func(args []RuntimeVal, env *Environment, pos Pos, _ *Interpreter) RuntimeVal {
		if len(args) < 2 {
			env.throwError([]string{"#_relative_path_to_file expects 2 arguments of type (string, string)", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		file, ok := args[0].(*StringVal)
		if !ok {
			env.throwError([]string{"#_relative_path_to_file expects it's 1st argument to be of type (string)", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		target, ok := args[1].(*StringVal)
		if !ok {
			env.throwError([]string{"#_relative_path_to_file expects it's 2nd argument to be of type (string)", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		return MK_STRING(RelativePathToFile(file.value, target.value))
	}))
	macros.set("#_stdin_prompt", MK_MACRO("#_stdin_prompt", func(args []RuntimeVal, env *Environment, pos Pos, _ *Interpreter) RuntimeVal {
		if len(args) < 2 {
			env.throwError([]string{"#_stdin_prompt expects 1 or 2 arguments of type (string, string)", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		message, ok := args[0].(*StringVal)
		if !ok {
			env.throwError([]string{"#_stdin_prompt expects it's 1st argument to be of type (string)", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		arg2 := args[1]
		input, err := Prompt(message.value, "")
		if err != nil {
			return arg2
		}
		return MK_STRING(input)
	}))
	macros.set("#_run_as_script", MK_MACRO("#_run_as_script", func(args []RuntimeVal, env *Environment, pos Pos, _ *Interpreter) RuntimeVal {
		if len(args) < 1 {
			env.throwError([]string{"#_run_as_script expects 1 argument of type (string)", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		path, ok := args[0].(*StringVal)
		if !ok {
			env.throwError([]string{"#_run_as_script expects it's 1st argument to be of type (string)", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		RunScript(path.value, env)
		return undefined
	}))
	macros.set("#_start_repl", MK_MACRO("#_start_repl", func(_ []RuntimeVal, _ *Environment, _ Pos, _ *Interpreter) RuntimeVal {
		REPL()
		return undefined
	}))
}

func createHttpHeaderObject(header http.Header, r *Interpreter) RuntimeVal {
	props := NewMap[string, RuntimeVal]()

	props.set("add", MK_MACRO("add", func(args []RuntimeVal, env *Environment, pos Pos, r *Interpreter) RuntimeVal {
		if len(args) < 2 {
			env.throwError([]string{"http.Header.add expects 2 arguments of type (string, string)", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		k, ok := args[0].(*StringVal)
		if !ok {
			env.throwError([]string{"http.Header.add expects it's 1st argument to be of type (string)", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		v, ok := args[1].(*StringVal)
		if !ok {
			env.throwError([]string{"http.Header.add expects it's 2nd argument to be of type (string)", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		header.Add(k.value, v.value)
		return undefined
	}))

	props.set("writer", MK_MACRO("writer", func(args []RuntimeVal, env *Environment, pos Pos, r *Interpreter) RuntimeVal {
		if len(args) < 1 {
			env.throwError([]string{"http.Header.writer expects 1 argument of type (raw [io writer])", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		k, ok := args[0].(*RawVal[io.Writer])
		if !ok {
			env.throwError([]string{"http.Header.writer expects it's 1st argument to be of type (raw [io writer])", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		err := header.Write(k.value)
		if err != nil {
			env.throwValue(MK_STRING(err.Error()), r)
		}
		return undefined
	}))

	props.set("set", MK_MACRO("set", func(args []RuntimeVal, env *Environment, pos Pos, r *Interpreter) RuntimeVal {
		if len(args) < 2 {
			env.throwError([]string{"http.Header.set expects 2 arguments of type (string, string)", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		k, ok := args[0].(*StringVal)
		if !ok {
			env.throwError([]string{"http.Header.set expects it's 1st argument to be of type (string)", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		v, ok := args[1].(*StringVal)
		if !ok {
			env.throwError([]string{"http.Header.set expects it's 1st argument to be of type (string)", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		header.Set(k.value, v.value)
		return undefined
	}))

	props.set("values", MK_MACRO("values", func(args []RuntimeVal, env *Environment, pos Pos, r *Interpreter) RuntimeVal {
		if len(args) < 1 {
			env.throwError([]string{"http.Header.values expects 1 argument of type (string)", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		k, ok := args[0].(*StringVal)
		if !ok {
			env.throwError([]string{"http.Header.values expects it's 1st argument to be of type (string)", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		values := header.Values(k.value)
		array := MK_ARRAY()
		for i := 0; i < len(values); i++ {
			array.Push(MK_STRING(values[i]))
		}
		return array
	}))

	props.set("remove", MK_MACRO("remove", func(args []RuntimeVal, env *Environment, pos Pos, r *Interpreter) RuntimeVal {
		if len(args) < 1 {
			env.throwError([]string{"http.Header.remove expects 1 argument of type (string)", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		k, ok := args[0].(*StringVal)
		if !ok {
			env.throwError([]string{"http.Header.remove expects it's 1st argument to be of type (string)", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		header.Del(k.value)
		return undefined
	}))

	props.set("get", MK_MACRO("get", func(args []RuntimeVal, env *Environment, pos Pos, r *Interpreter) RuntimeVal {
		if len(args) < 1 {
			env.throwError([]string{"http.Header.get expects 1 argument of type (string)", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		k, ok := args[0].(*StringVal)
		if !ok {
			env.throwError([]string{"http.Header.get expects it's 1st argument to be of type (string)", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		return MK_STRING(header.Get(k.value))
	}))

	props.set("clone", MK_MACRO("clone", func(args []RuntimeVal, env *Environment, pos Pos, r *Interpreter) RuntimeVal {
		return createHttpHeaderObject(header.Clone(), r)
	}))

	object := NewMap[RuntimeVal, string]()
	props.forEach(func(key string, value RuntimeVal) {
		ml := GenerateRadix(16)
		Memory.set(ml, value)
		object.set(MK_STRING(key), ml)
	})
	return MK_OBJECT(object, nil, r)
}
