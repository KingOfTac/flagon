import { type Terminal } from '@xterm/xterm';
import {
  c,
  CCliAnsiColor,
  CCliCommand,
  CCliCommandGroup,
  type CCliSubcommand,
  navigateCCliTree,
} from '@carnesen/cli';

interface SplitCommandLineResult {
  args: string[];
  quoteChar: string;
}

function splitCommandLine(commandLine: string): SplitCommandLineResult {
  const args: string[] = [];
  let currentArg: string | undefined;
  let quoteChar = '';

  function appendToCurrentArg(char: string) {
    currentArg = (currentArg ?? '') + char;
  }

  for (const char of commandLine) {
    switch (char) {
      case '"':
      case "'":
        if (quoteChar) {
          if (char === quoteChar) {
            quoteChar = '';
          } else {
            appendToCurrentArg(char);
          }
        } else {
          quoteChar = char;
        }
        break;

      case ' ':
        if (quoteChar) {
          appendToCurrentArg(char);
        } else if (typeof currentArg === 'string') {
          args.push(currentArg);
          currentArg = undefined;
        }
        break;

      default:
        appendToCurrentArg(char);
        break;
    }
  }

  if (currentArg) {
    args.push(currentArg);
  }

  return {
    args,
    quoteChar,
  };
}

// TODO: cleanup minified code and add types
export class CommandLine {
  private line: string;
  private currentIndex: number;

  constructor(line: string = '', index?: number) {
    this.line = line;
    this.currentIndex =
      typeof index === 'number' ? this.indexInRange(index) : this.line.length;
  }

  public splitIntoArgs(line: string = this.line) {
    const { args, quoteChar } = splitCommandLine(line);
    return {
      args,
      singleQuoted: quoteChar === "'",
      doubleQuoted: quoteChar === '"',
    };
  }

  public indexInRange(index: number): number {
    return Math.min(Math.max(0, index), this.line.length);
  }

  public splitIntoArgsAndSearch() {
    const lineUpToCursor = this.line.substring(0, this.currentIndex);
    const { args, singleQuoted, doubleQuoted } =
      this.splitIntoArgs(lineUpToCursor);

    let search = '';
    let completedArgs: string[];

    if (lineUpToCursor.endsWith(' ')) {
      completedArgs = args;
    } else {
      search = args.at(-1) ?? '';
      completedArgs = args.slice(0, -1);
    }

    return {
      args: completedArgs,
      search,
      singleQuoted,
      doubleQuoted,
    };
  }

  public delete(): boolean {
    if (this.currentIndex === 0) {
      return false;
    }
    this.line =
      this.line.substring(0, this.currentIndex - 1) +
      this.line.substring(this.currentIndex);
    this.currentIndex--;
    return true;
  }

  public insert(text: string): string {
    if (text.length === 0) {
      return '';
    }
    const before = this.line.substring(0, this.currentIndex);
    const after = this.line.substring(this.currentIndex);
    this.line = `${before}${text}${after}`;
    this.currentIndex += text.length;
    return `${text}${after}${'\b'.repeat(after.length)}`;
  }

  public next(): boolean {
    if (this.currentIndex < this.line.length) {
      this.currentIndex++;
      return true;
    }
    return false;
  }

  public previous(): boolean {
    if (this.currentIndex > 0) {
      this.currentIndex--;
      return true;
    }
    return false;
  }

  public reset(): void {
    this.line = '';
    this.currentIndex = 0;
  }

  public setValue(value: string): string {
    const lengthDiff = value.length - this.line.length;
    let sequence = `\b`.repeat(this.currentIndex) + value;

    if (lengthDiff < 0) {
      const padding = ' '.repeat(-lengthDiff);
      const backspaces = '\b'.repeat(-lengthDiff);
      sequence += padding + backspaces;
    }

    this.line = value;
    this.currentIndex = this.line.length;

    const backspacesToCursor = '\b'.repeat(value.length - this.currentIndex);
    sequence += backspacesToCursor;

    return sequence;
  }

  public sequence(): string {
    return `${this.line}${'\b'.repeat(this.line.length - this.currentIndex)}`;
  }

  public value(): string {
    return this.line;
  }
}

export class CommandHistory {
  private history: string[] = [];
  private index = 0;

  constructor(initialHistory: string[] = [], currentLine = '') {
    this.setHistory(...initialHistory.map((line) => line.trim()));
    this.current(currentLine);
  }

  public current(line?: string): string {
    if (typeof line === 'string') {
      this.history[this.index] = line;
    }
    return this.history[this.index];
  }

  public submit(line: string): void {
    const trimmedLine = line.trim();
    this.current(trimmedLine);

    const newHistory = this.history.filter((h) => h !== trimmedLine);
    this.setHistory(...newHistory);

    if (this.current() !== trimmedLine) {
      this.setHistory(...this.history, trimmedLine);
    }
  }

  public previous(currentLine: string): string {
    this.current(currentLine);
    if (this.index > 0) {
      this.index -= 1;
    }
    return this.current();
  }

  public next(currentLine: string): string {
    this.current(currentLine);
    if (this.index < this.indexOfLastLine()) {
      this.index += 1;
    }
    return this.current();
  }

  public list(): string[] {
    return this.history.filter((line) => line.length > 0);
  }

  public indexOfLastLine(): number {
    return this.history.length - 1;
  }

  public setHistory(...lines: string[]): void {
    this.history = lines.filter((line) => line.length > 0);
    this.history.push('');
    this.index = this.indexOfLastLine();
  }
}

function C(e, t) {
  let r = e.filter((e) => e.startsWith(t)).map((e) => e.slice(t.length));
  switch (r.length) {
    case 0:
      return [];
    case 1:
      return [''.concat(r[0], ' ')];
    default: {
      let e = (function e(t) {
        let r =
          arguments.length > 1 && void 0 !== arguments[1] ? arguments[1] : '';
        if (0 === t.length || r.length >= t[0].length) return r;
        let o = t[0].substring(0, r.length + 1);
        for (let e = 1; e < t.length; e += 1)
          if (t[e].substring(0, r.length + 1) !== o) return r;
        // @ts-ignore
        return e(t, o);
      })(r);
      if (e.length > 0) return [e];
      return r;
    }
  }
}

function f(e, t, r) {
  return e ? C(e._suggest(t, r), r) : [];
}

type ReplOptions = {
  description: string;
  terminal: Terminal;
  submit: any;
  subcommands: CCliSubcommand[];
};

const color = CCliAnsiColor.create();
const promptStart = `${color.green('$')} `;

const promptWelcome = (terminal: Terminal) => {
  const br = Array(terminal.cols).fill('-').join('');
  return `${br}\n[?] Need help? Type 'help' to see available commands\n${br}`;
};

export class Repl {
  private terminal: Terminal;
  private submit: boolean;
  private commandLine: CommandLine = new CommandLine();
  private commandHistory: CommandHistory = new CommandHistory();
  private runningCommand: boolean = false;
  private settingCurrentLine: boolean = false;
  public root: CCliCommandGroup;

  constructor({ description, terminal, submit, subcommands }: ReplOptions) {
    this.terminal = terminal;
    this.submit = submit;

    const commands = [
      ...subcommands,
      c.command({
        name: 'clear',
        description: 'Clears the terminal',
        action: () => {
          this.terminal.clear();
          console.clear();
        },
      }),
      c.command({
        name: 'help',
        description: 'Lists the commands available for you to use',
        action: () => globalThis.runCommand('help')
      }),
    ];

    this.root = c.commandGroup({
      description,
      name: '',
      subcommands: commands,
    });
  }

  public start() {
    this.terminal.onKey((args) => this.handleKeyEvent(args));
    // this.terminal.focus();
    this.submit || this.consoleLog(promptWelcome(this.terminal));
    this.prompt();
    this.submit && this.runCurrentLine();
  }

  public setAndRunArgs(e: string[]) {
    this.setLine(e.join(' '));
    this.runCurrentLine();
  }

  public prompt() {
    this.terminal.write(`${promptStart}${this.commandLine.sequence()}`);
  }

  public handleKeyEvent({
    key,
    domEvent,
  }: {
    key: string;
    domEvent: KeyboardEvent;
  }) {
    if (this.settingCurrentLine || this.runningCommand) {
      return;
    }

    let noMod = !domEvent.altKey && !domEvent.ctrlKey && !domEvent.metaKey;
    domEvent.preventDefault();

    switch (domEvent.keyCode) {
      case 8:
        this.commandLine.delete() && this.terminal.write('\b \b');
        break;

      case 9:
        this.autoComplete();
        break;

      case 13:
        this.runCurrentLine();
        break;

      case 37:
        this.commandLine.previous() && this.terminal.write(key);
        break;

      case 39:
        this.commandLine.next() && this.terminal.write(key);
        break;

      case 38:
        this.setLine(this.commandHistory.previous(this.commandLine.value()));
        break;

      case 40:
        this.setLine(this.commandHistory.next(this.commandLine.value()));
        break;

      default:
        if (noMod) {
          this.addToLine(key);
        }
    }
  }

  public consoleLog(e) {
    let t;
    (t = (t =
      'string' == typeof e
        ? e.startsWith('\n')
          ? '\r'.concat(e)
          : e
        : 'string' == typeof e.stack
        ? e.stack
        : '').replace(/([^\r])\n/g, '$1\r\n')).endsWith('\r\n') ||
      (t += '\r\n');
    this.terminal.write(t);
  }

  public consoleError(e) {
    console.log(e);
    this.consoleLog(e);
  }

  public runCurrentLine() {
    const self = this;
    this.terminal.write('\r\n');
    this.commandHistory.submit(this.commandLine.value());

    let { args, singleQuoted, doubleQuoted } = this.commandLine.splitIntoArgs();

    if (singleQuoted || doubleQuoted) {
      this.consoleError(
        `Error: ${singleQuoted ? 'single' : 'double'} quotes are not balanced`
      );
      this.prompt();
      return;
    }

    let i = {
      logger: {
        error: function () {
          let i = arguments.length;
          let r = Array(i);

          for (let o = 0; o < i; o++) {
            r[o] = arguments[o];
          }
          self.consoleError(r[0]);
        },
        log: function () {
          let i = arguments.length;
          let r = Array(i);

          for (let o = 0; o < i; o++) {
            r[o] = arguments[o];
          }

          self.consoleLog(r[0]);
        },
      },
      columns: this.terminal.cols,
    };

    this.runningCommand = true;

    c.cli(this.root, i)
      .run(args)
      .then(() => {
        this.commandLine.reset();
        this.runningCommand = false;
        this.prompt();
      });
  }

  public setLine(e) {
    this.settingCurrentLine = true;
    let t = this.commandLine.setValue(e);
    this.terminal.write(t, () => {
      this.settingCurrentLine = false;
    });
  }

  public addToLine(e) {
    this.terminal.write(this.commandLine.insert(e));
  }

  public autoComplete() {
    let {
        args: e,
        search: t,
        singleQuoted: r,
        doubleQuoted: o,
      } = this.commandLine.splitIntoArgsAndSearch(),
      i = (function (e, t, r) {
        if (t.includes('--help')) return [];
        // @ts-ignore
        let o = (0, navigateCCliTree)(e, t);
        if (o.tree.current instanceof CCliCommandGroup)
          return o.args.length > 0
            ? []
            : C(
                [
                  ...o.tree.current.subcommands.map((e) => {
                    let { name: t } = e;
                    return t;
                  }),
                ],
                r
              );
        if (o.tree.current instanceof CCliCommand) {
          let e = o.tree.current,
            t = [
              ...(e.namedArgGroups
                ? Object.keys(e.namedArgGroups).map((e) => '--'.concat(e))
                : []),
              '--help',
            ],
            i = o.args.slice(-1)[0];
          if ('--' === i) return f(o.tree.current.doubleDashArgGroup, [], r);
          if (o.args.includes('--')) return [];
          if ('-' === r || '--' === r) {
            let o = [];
            return e.doubleDashArgGroup && o.push('--'), o.push(...t), C(o, r);
          }
          if (r.startsWith('--')) return C(t, r);
          if (0 === o.args.length) {
            let { positionalArgGroup: o } = e,
              i = f(o, [], r);
            if (o && !o.optional) return i;
            let n = [];
            return '' === r
              ? (e.doubleDashArgGroup && n.push('--'),
                n.push(...t),
                [...n, ...i])
              : i;
          }
          return i.startsWith('--') && o.tree.current.namedArgGroups
            ? f(o.tree.current.namedArgGroups[i.slice(2)], [], r)
            : [];
        }
        throw Error('Unexpected kind');
      })(this.root, e, t);
    switch (i.length) {
      case 0:
        return;
      case 1: {
        let e = i[0];
        if (i[0].endsWith(' ')) {
          let t = e.substring(0, i[0].length - 1);
          r ? (e = ''.concat(t, "' ")) : o && (e = ''.concat(t, '" '));
        }
        this.addToLine(e);
        break;
      }
      default:
        this.consoleLog(''),
          i.forEach((e) => {
            this.consoleLog(''.concat('    ').concat(t).concat(e));
          }),
          this.prompt();
    }
  }
}
