import '@neutron-ui/neutron/bundle.css';
import './styles.css';

const luaEditor = document.getElementById('lua') as HTMLTextAreaElement;
const consoleWindow = document.getElementById('console') as HTMLTextAreaElement;

luaEditor.addEventListener('input', () => {
  const code = luaEditor.value;
  const preview = document.getElementById('preview') as HTMLElement;
  preview.textContent = code;
});

let wasmLoaded = false;
const go = new Go();
WebAssembly.instantiateStreaming(fetch('/flagon.wasm'), go.importObject)
  .then((result) => {
    go.run(result.instance);
    wasmLoaded = true;
    log('WASM module loaded and Go runtime started');
  });

function log(...data: any[]) {
  console.log(...data);
  consoleWindow.value += data.join(' ') + '\n';
}

const tabs = document.getElementById('tabs') as HTMLElement;