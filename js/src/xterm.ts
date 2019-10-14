import { Terminal } from 'xterm';
import { FitAddon } from 'xterm-addon-fit';
import { lib } from "libapps"

const terminal = new Terminal()
terminal.loadAddon(new FitAddon());

export class Xterm {
    elem: HTMLElement;
    term: Terminal;
    resizeListener: () => void;
    decoder: lib.UTF8Decoder;

    message: HTMLElement;
    messageTimeout: number;
    messageTimer: number;


    constructor(elem: HTMLElement) {
        this.elem = elem;
        this.term = new Terminal();
        const fitAddon = new FitAddon();
        this.term.loadAddon(fitAddon);

        if (elem.ownerDocument) {
            this.message = elem.ownerDocument.createElement("div") ;
        }
        this.message.className = "xterm-overlay";
        this.messageTimeout = 2000;

        this.resizeListener = () => {
            fitAddon.fit();
            this.term.scrollToBottom();
            this.showMessage(String(this.term.cols) + "x" + String(this.term.rows), this.messageTimeout);
        };

        (<any>this).term.on("open", () => {
            this.resizeListener();
            window.addEventListener("resize", () => { this.resizeListener(); });
        });

        (<any>this).term.open(elem, true);

        this.decoder = new lib.UTF8Decoder()
    };

    info(): { columns: number, rows: number } {
        return { columns: this.term.cols, rows: this.term.rows };
    };

    output(data: string) {
        this.term.write(this.decoder.decode(data));
    };

    showMessage(message: string, timeout: number) {
        this.message.textContent = message;
        this.elem.appendChild(this.message);

        if (this.messageTimer) {
            clearTimeout(this.messageTimer);
        }
        if (timeout > 0) {
            this.messageTimer = setTimeout(() => {
                this.elem.removeChild(this.message);
            }, timeout);
        }
    };

    removeMessage(): void {
        if (this.message.parentNode == this.elem) {
            this.elem.removeChild(this.message);
        }
    }

    setWindowTitle(title: string) {
        document.title = title;
    };

    setPreferences(value: object) {
    };

    onInput(callback: (input: string) => void) {
        (<any>this).term.on("data", (data) => {
            callback(data);
        });

    };

    onResize(callback: (colmuns: number, rows: number) => void) {
        (<any>this).term.on("resize", (data) => {
            callback(data.cols, data.rows);
        });
    };

    deactivate(): void {
        (<any>this).term.off("data");
        (<any>this).term.off("resize");
        this.term.blur();
    }

    reset(): void {
        this.removeMessage();
        this.term.clear();
    }

    close(): void {
        window.removeEventListener("resize", this.resizeListener);
        (<any>this).term.destroy();
    }
}
