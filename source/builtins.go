package main

import (
	"fmt"
	"runtime"
	"slices"
)

func (r *Interpreter) IsMemoryHigh() bool {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.HeapInuse > r.MemThreshold // Or use m.Alloc, etc.
}

type Task struct {
	macro *Macro
	args  []RuntimeVal
	env   *Environment
	pos   Pos
	r     *Interpreter
}

type MicroTaskQueue struct {
	queue  []Task
	length int
}

func (q *MicroTaskQueue) queueMicroTask(task Task) {
	q.queue = append(q.queue, task)
	q.length = len(q.queue)
}

func (q *MicroTaskQueue) execCurrentTask() {
	if len(q.queue) == 0 {
		return
	}
	index := len(q.queue) - 1
	task := q.queue[index]
	q.queue = slices.Delete(q.queue, index, index+1)
	// fmt.Printf("\x1b[34mTask\x1b[0m\r\nmacro: %+v, args: %+v\r\n", task.macro, task.args)
	task.macro.call(task.args, task.env, task.pos, task.r)
	q.length = len(q.queue)
}

var promise_mem_loc = GenerateRadix(16)

type Callable interface {
	RuntimeVal
	Call(env *Environment, args []RuntimeVal, r *Interpreter, pos Pos) RuntimeVal
}

func Promise(fn Callable, fn_args []RuntimeVal, r *Interpreter, env *Environment, pos Pos) *Instance {
	var declEnv *Environment
	switch fn := fn.(type) {
	case *FunctionVal:
		declEnv = fn.declEnv
	case *Macro:
		declEnv = env
	}
	PromiseExecutorWrapper := MK_MACRO("#_promise_exec_wrapper", func(args []RuntimeVal, env *Environment, pos Pos, r *Interpreter) RuntimeVal {
		args_len := len(args)
		if args_len < 2 {
			env.throwError([]string{"#_promise_exec_wrapper expects 2 arguments, but it was given", fmt.Sprint(args_len), SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		exec := args[0].(Callable)
		class := args[1].(*Instance)

		resolve, ok := GetInstanceMember(class, "resolve").(*Macro)
		// reject, ok := GetNativeMember(class, "reject").(*Macro)
		switch exec := exec.(type) {
		case *FunctionVal:
			// keep this line or else... you want to cry
			exec.async = false
		}
		value := exec.Call(env, fn_args, r, pos)
		if ok {
			resolve.call([]RuntimeVal{value, class}, env, pos, r)
		}
		return value
	})

	var class *NativeClass
	var instance *Instance

	props := NewMap[string, string]()
	// then
	thenCallback := GenerateRadix(16)
	Memory.set(thenCallback, undefined)
	props.set("thenCallback", thenCallback)
	// catch
	catchCallback := GenerateRadix(16)
	Memory.set(catchCallback, undefined)
	props.set("catchCallback", catchCallback)
	// finally
	finallyCallback := GenerateRadix(16)
	Memory.set(finallyCallback, undefined)
	props.set("finallyCallback", finallyCallback)

	resolve := MK_MACRO(
		"resolve",
		func(args []RuntimeVal, env *Environment, pos Pos, r *Interpreter) RuntimeVal {
			var value RuntimeVal
			if len(args) < 2 {
				return undefined
			}
			value = args[0]
			promise, ok := args[1].(*Instance)
			if ok {
				then := GetInstanceMember(promise, "thenCallback")
				args := []RuntimeVal{value}
				switch then := then.(type) {
				case *Macro:
					then.call(args, env, pos, r)
				case *FunctionVal:
					then.Call(env, args, r, pos)
				}
			}
			return value
		},
	)
	then := MK_MACRO("then", func(args []RuntimeVal, env *Environment, pos Pos, r *Interpreter) RuntimeVal {
		var callback RuntimeVal
		args_len := len(args)
		if args_len < 1 {
			env.throwError([]string{"Promise.then expects 1 argument (callback), but it was given", fmt.Sprint(args_len), SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
			return undefined
		}
		callback = args[0]
		Memory.set(thenCallback, callback)
		return instance
	})
	catch := MK_MACRO("catch", func(args []RuntimeVal, env *Environment, pos Pos, r *Interpreter) RuntimeVal {
		var callback RuntimeVal
		args_len := len(args)
		if args_len < 2 {
			callback = undefined
			return callback
		}
		callback = args[0]
		promise, ok := args[1].(*Instance)
		if !ok {
			env.throwError([]string{"Promise.catch expects 2 arguments (callback, Promise), but it was given", fmt.Sprint(args_len), SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		Memory.set(catchCallback, callback)
		return promise
	})
	finally := MK_MACRO("finally", func(args []RuntimeVal, env *Environment, pos Pos, r *Interpreter) RuntimeVal {
		var callback RuntimeVal
		args_len := len(args)
		if args_len < 2 {
			callback = undefined
			return callback
		}
		callback = args[0]
		promise, ok := args[1].(*Instance)
		if !ok {
			env.throwError([]string{"Promise.finally expects 2 arguments (callback, Promise), but it was given", fmt.Sprint(args_len), SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		Memory.set(finallyCallback, callback)
		return promise
	})
	methods := []*Macro{
		resolve, then, catch, finally,
	}
	constructor := MK_MACRO("constructor", func(args []RuntimeVal, env *Environment, pos Pos, r *Interpreter) RuntimeVal {
		if len(args) == 0 {
			env.throwError([]string{"Promise expects one argument, but it was given none", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		callback := args[0]
		vt := ValueType(callback)
		if !is_value(vt, "function", "macro") {
			env.throwError([]string{"Promise expects an argument of type 'function', but it was given one of", vt, SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")})
		}
		r.microTaskQueue.queueMicroTask(Task{
			macro: PromiseExecutorWrapper,
			args:  []RuntimeVal{callback, instance},
			env:   declEnv,
			pos:   pos,
			r:     r,
		})
		return undefined
	})
	class = MK_NT_CLASS("Promise", constructor, props, methods, declEnv, "")
	Memory.set(promise_mem_loc, class)
	instance_props := NewMap[RuntimeVal, string]()
	class.properties.forEach(func(key, value string) {
		instance_props.set(MK_STRING(key), value)
	})
	for _, method := range class.methods {
		ml := GenerateRadix(16)
		Memory.set(ml, method)
		instance_props.set(MK_STRING(method.name), ml)
	}
	instance = MK_INSTANCE(class.name, promise_mem_loc, instance_props, GetUDRef(r), declEnv, r)
	class.ctor.call([]RuntimeVal{fn}, declEnv, pos, r)
	return instance
}

func Instantiate(ml string, args []RuntimeVal, env *Environment, r *Interpreter, pos Pos) *Instance {
	class, ok := Memory.get(ml).(*NativeClass)
	props := NewMap[RuntimeVal, string]()
	if ok {
		var _default string = GetUDRef(r)
		class.properties.forEach(func(key, value string) {
			props.set(MK_STRING(key), value)
			_default = value
		})
		for _, method := range class.methods {
			ml := GenerateRadix(16)
			Memory.set(ml, method)
			props.set(MK_STRING(method.name), ml)
		}
		class.ctor.call(args, env, pos, r)
		return MK_INSTANCE(class.name, ml, props, _default, env, r)
	}
	env.ThrowTypeError("type", ValueType(class), "is not a class and is not constructable.")
	return nil
}

func GetInstanceMember(class *Instance, member string) RuntimeVal {
	prop := MK_STRING(member)
	ml := class.properties.get(prop)
	if len(ml) == 0 {
		p := class.prototype
		ml = GetPropMlFromProto(prop, p)
		if len(ml) == 0 {
			return MK_UD()
		}
	}
	return Memory.get(ml)
}
