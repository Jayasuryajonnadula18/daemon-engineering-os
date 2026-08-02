const vscode = require('vscode');
const cp = require('child_process');
const path = require('path');
const fs = require('fs');

let masterPasswordCache = "";
let statusBarItem;
let outputChannels = {};

function getBinaryPath() {
    // 1. Explicit env override
    if (process.env.DAEMON_PATH && fs.existsSync(process.env.DAEMON_PATH)) {
        return process.env.DAEMON_PATH;
    }
    // 2. Global WindowsApps install
    const globalPath = path.join(process.env.LOCALAPPDATA || '', 'Microsoft', 'WindowsApps', 'daemon.exe');
    if (fs.existsSync(globalPath)) return globalPath;
    // 3. Relative to extension directory
    const relPath = path.join(__dirname, '..', 'daemon', 'daemon.exe');
    if (fs.existsSync(relPath)) return relPath;
    // 4. Hardcoded dev path
    return 'c:/Users/MAHESH/OneDrive/Desktop/Daemon CLI/daemon/daemon.exe';
}

function getOutputChannel(name) {
    if (!outputChannels[name]) {
        outputChannels[name] = vscode.window.createOutputChannel(name);
    }
    return outputChannels[name];
}

function getPassword(callback) {
    if (masterPasswordCache) {
        callback(masterPasswordCache);
        return;
    }
    if (process.env.DAEMON_PASSWORD) {
        masterPasswordCache = process.env.DAEMON_PASSWORD;
        updateStatusBar('active');
        callback(masterPasswordCache);
        return;
    }
    // Query daemon token dynamically from OS Keyring via binary
    const binaryPath = getBinaryPath();
    cp.exec(`"${binaryPath}" token`, (err, stdout) => {
        if (!err && stdout.includes("Token:")) {
            const match = stdout.match(/Token:\s+([a-f0-9]+)/i);
            if (match && match[1]) {
                masterPasswordCache = match[1];
                updateStatusBar('active');
                callback(masterPasswordCache);
                return;
            }
        }
        // Fallback: prompt user if keyring fetch fails
        vscode.window.showInputBox({
            prompt: "Daemon Master Password / Token",
            password: true,
            placeHolder: "Enter your OS Keyring master token (run 'daemon token' to view)"
        }).then(pwd => {
            if (pwd) {
                masterPasswordCache = pwd;
                updateStatusBar('active');
                callback(pwd);
            } else {
                vscode.window.showErrorMessage("Daemon: Unauthorized — token required.");
            }
        });
    });
}

function runDaemonCommand(channelName, args, password, inputPrompt) {
    const channel = getOutputChannel(channelName);
    channel.show(true);
    channel.appendLine(`\n>>> daemon ${args.join(' ')}\n`);

    const binaryPath = getBinaryPath();
    const fullArgs = [...args, '--password', password];
    const proc = cp.spawn(binaryPath, fullArgs, {
        env: { ...process.env, DAEMON_PASSWORD: password }
    });

    proc.stdout.on('data', d => channel.append(d.toString()));
    proc.stderr.on('data', d => channel.append(d.toString()));
    proc.on('close', code => {
        channel.appendLine(`\n[Daemon exited with code ${code}]`);
    });
}

function updateStatusBar(state) {
    if (!statusBarItem) return;
    if (state === 'active') {
        statusBarItem.text = '$(pulse) Daemon: Active';
        statusBarItem.tooltip = 'Daemon Engineering OS is running. Click to open Mission Control.';
        statusBarItem.command = 'daemon.mission';
        statusBarItem.backgroundColor = undefined;
    } else {
        statusBarItem.text = '$(circle-slash) Daemon: Locked';
        statusBarItem.tooltip = 'Click any Daemon command to authenticate.';
        statusBarItem.command = 'daemon.advise';
    }
    statusBarItem.show();
}

function activate(context) {
    console.log('Daemon Engineering OS extension activated');

    // Status bar
    statusBarItem = vscode.window.createStatusBarItem(vscode.StatusBarAlignment.Left, 100);
    updateStatusBar('locked');
    context.subscriptions.push(statusBarItem);

    // Command definitions: [commandId, channelName, cliArgs, requiresInput]
    const commands = [
        { id: 'daemon.advise',   channel: 'Daemon Advisor',       args: ['advise'],          input: null },
        { id: 'daemon.maintain', channel: 'Daemon Maintenance',    args: ['maintain'],        input: null },
        { id: 'daemon.doctor',   channel: 'Daemon Doctor',         args: ['doctor'],          input: null },
        { id: 'daemon.sync',     channel: 'Daemon Sync',           args: ['sync'],            input: null },
        { id: 'daemon.graph',    channel: 'Daemon Graph',          args: ['graph'],           input: null },
        { id: 'daemon.version',  channel: 'Daemon Version',        args: ['version'],         input: null },
        { id: 'daemon.config',   channel: 'Daemon Config',         args: ['config', 'view'],  input: null },
        { id: 'daemon.replay',   channel: 'Daemon Replay',         args: ['replay', 'today'], input: null },
        { id: 'daemon.fix',      channel: 'Daemon Fix',            args: ['fix', '--dry-run'],input: null },
    ];

    commands.forEach(({ id, channel, args }) => {
        const disposable = vscode.commands.registerCommand(id, () => {
            getPassword(pwd => runDaemonCommand(channel, args, pwd));
        });
        context.subscriptions.push(disposable);
    });

    // daemon.plan — requires intent input
    context.subscriptions.push(vscode.commands.registerCommand('daemon.plan', () => {
        getPassword(pwd => {
            vscode.window.showInputBox({ prompt: 'What is your intent? (e.g. Deploy orders-api to production)' }).then(intent => {
                if (!intent) return;
                runDaemonCommand('Daemon Planner', ['plan', intent], pwd);
            });
        });
    }));

    // daemon.deploy — requires service input + strategy picker
    context.subscriptions.push(vscode.commands.registerCommand('daemon.deploy', () => {
        getPassword(pwd => {
            vscode.window.showInputBox({ prompt: 'Service to deploy (e.g. orders-api)' }).then(service => {
                if (!service) return;
                vscode.window.showQuickPick(['standard', 'canary', 'blue-green'], { placeHolder: 'Deployment strategy' }).then(strategy => {
                    runDaemonCommand('Daemon Deploy', ['deploy', service, '--strategy', strategy || 'standard'], pwd);
                });
            });
        });
    }));

    // daemon.mission — launch Mission Control in browser
    context.subscriptions.push(vscode.commands.registerCommand('daemon.mission', () => {
        getPassword(pwd => {
            const channel = getOutputChannel('Daemon Mission Control');
            channel.show(true);
            channel.appendLine('\n>>> Launching Mission Control...\n');
            vscode.env.openExternal(vscode.Uri.parse('http://127.0.0.1:8081'));
            channel.appendLine('✔ Opened Mission Control at http://127.0.0.1:8081');
            channel.appendLine('  If the dashboard is not running, execute: daemon mission');
        });
    }));
}

function deactivate() {
    Object.values(outputChannels).forEach(ch => ch.dispose());
}

module.exports = { activate, deactivate };
