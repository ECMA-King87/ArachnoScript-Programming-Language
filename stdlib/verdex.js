//@ts-check

class _VDX_Element1 {
  /**
   * @type {HTMLElement}
   */
  #element = null;
  /**
   * @param {HTMLElement} element
   */
  constructor(element) {
    this.#element = element;
  }
  /**
   * @param {string} text
   * @returns
   */
  text(text) {
    return this.#element.innerText = text ?? this.#element.innerText;
  }
  /**
   * @param {string} html
   * @returns
   */
  html(html) {
    return this.#element.innerHTML = html ?? this.#element.innerHTML;
  }
}

/**
 * @param {string} selector
 * @returns {_VDX_Element1}
 */
function $(selector) {
  if (typeof selector != "string") {
    throw "$: argument must be a string.";
  }
  const el = document.querySelector(selector);
  if (el == null) {
    throw "$: element " + selector + " does not exist.";
  }
  return new _VDX_Element1(el);
}
const __vx_nodes = new Map();
/**
 * @param {string} id
 * @param {string} getter
 */
function __vdx_bind(id, getter) {
  __vx_nodes.set(id, getter);
}

function __vdx_update() {
  for (const [id, getter] of __vx_nodes) {
    const el = $(`[data-vx="${id}"]`);
    el.text(getter());
  }
}
