import { useState } from "react";

const memoryState = {};

export default function(key, initialState) {
  const [state, setState] = useState(() => {
    const hasMemoryValue = Object.prototype.hasOwnProperty.call(memoryState, key);
    if (hasMemoryValue) {
      return memoryState[key]
    } else {
      return typeof initialState === 'function' ? initialState() : initialState;
    }
  });

  function onChange(nextState) {
    memoryState[key] = nextState;
    setState(nextState);
  }

  return [state, onChange];
}