import '@xterm/xterm/css/xterm.css';
import 'highlight.js/styles/base16/bright.css';
import '@neutron-ui/neutron/bundle.css';
import './styles.css';

import { Terminal } from '@xterm/xterm';
import { Repl } from './repl';
import hljs from 'highlight.js/lib/core';
import lua from 'highlight.js/lib/languages/lua';
import { c } from '@carnesen/cli';

hljs.registerLanguage('lua', lua);

const luaEditor = document.getElementById('lua') as HTMLTextAreaElement;
const consoleWindow = document.getElementById('console') as HTMLTextAreaElement;
const terminal = document.getElementById('cli') as HTMLElement;
const preview = document.getElementById('preview') as HTMLElement;
const loadLuaButton = document.getElementById('load-lua') as HTMLButtonElement;

const term = new Terminal({
  cursorBlink: true,
  fontFamily: 'ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas,\'Liberation Mono\', \'Courier New\', monospace',
  fontSize: 14,
  theme: {}
});

const repl = new Repl({
  description: 'Flagon CLI',
  subcommands: [],
  submit: false,
  terminal: term,
});

term.open(terminal);
repl.start();

if (luaEditor.innerHTML.trim().length > 0) {
  const code = luaEditor.value;
  const highlightedCode = hljs.highlight(code, { language: 'lua' }).value;
  preview.innerHTML = `<pre><code class="hljs lua">${highlightedCode}</code></pre>`;
}

luaEditor.addEventListener('input', () => {
  const code = luaEditor.value;
  const highlightedCode = hljs.highlight(code, { language: 'lua' }).value;
  preview.innerHTML = `<pre><code class="hljs lua">${highlightedCode}</code></pre>`;
});

loadLuaButton.addEventListener('click', () => {
  const code = luaEditor.value;
  const result = globalThis.loadLuaPlugin(code);

  try {
    const command = JSON.parse(result);
  
    log('Loaded Lua plugin command:', command.name);
  
    repl.root.subcommands.push(c.command({
      name: command.name,
      description: command.description,
      action: () => globalThis.runCommand(command.name)
    }))
  } catch (err) {
    log(result.error);
  }
});

let wasmLoaded = false;
const go = new Go();
WebAssembly.instantiateStreaming(fetch('/flagon.wasm'), go.importObject)
  .then((result) => {
    go.run(result.instance);
    wasmLoaded = true;
    log('WASM module loaded and Go runtime started');
  }).catch((err) => {
    log('Error loading WASM module:', err);
  });

function log(...data: any[]) {
  console.log(...data);
  consoleWindow.value += data.join(' ') + '\n';
}

const tabs = document.getElementById('tabs') as HTMLElement;