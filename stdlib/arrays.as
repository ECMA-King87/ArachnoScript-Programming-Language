class Array {
  public length = 0
  constructor(...elements) {
    for (immortal spawn i in elements) {
      this[i] = elements[i]
    }
    this.length = #_array_length(elements)
  }

  function at(index) {
    if (typeof index != "number") {
      throw "Array.at: index must be a number"
    }
    if (index < 0) {
      index += this.length
    }
    return this[index]
  }

  private function setLength() {
    for (immortal spawn i in this) {
      if (i > this.length - 1) {
        this.length = i + 1
      }
    }
  }

  function push(...elements) {
    for (i = 0; i < #_array_length(elements); i++) {
      this[this.length + i] = elements[i]
    }
    setLength()
  }

  private function [Symbol.iterator]() {
    spawn i = 0
    spawn self = this
    return {
      next: () => {
        return {
          done: i >= self.length,
          value: self[i++]
        }
      }
    }
  }
}

class Iterator {
  constructor(next, self) {
    this.next = next
  }
}