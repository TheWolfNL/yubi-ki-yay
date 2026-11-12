import {SettingsService} from "../bindings/github.com/TheWolfNL/yubi-ki-yay";
import type {Config} from "../bindings/github.com/TheWolfNL/yubi-ki-yay/internal/config";

export function initSettings() {
    const form = document.getElementById("settings") as HTMLFormElement;
    const soundSelect = form.elements.namedItem("soundName") as HTMLSelectElement;
    const trayModeSelect = form.elements.namedItem("trayMode") as HTMLSelectElement;
    const bannerPosSelect = form.elements.namedItem("bannerPosition") as HTMLSelectElement;
    const status = document.getElementById("status")!;
    let statusTimer: ReturnType<typeof setTimeout>;

    const toggles = ["banner", "sound", "terminalLine", "trayFlash"] as const;

    function checkbox(name: (typeof toggles)[number]): HTMLInputElement {
        return form.elements.namedItem(name) as HTMLInputElement;
    }

    function flash(message: string) {
        status.textContent = message;
        clearTimeout(statusTimer);
        statusTimer = setTimeout(() => (status.textContent = ""), 2000);
    }

    function read(): Config {
        return {
            banner: checkbox("banner").checked,
            bannerPosition: bannerPosSelect.value,
            sound: checkbox("sound").checked,
            soundName: soundSelect.value,
            terminalLine: checkbox("terminalLine").checked,
            trayFlash: checkbox("trayFlash").checked,
            trayMode: trayModeSelect.value,
        };
    }

    const removeButton = document.getElementById("remove-sound") as HTMLButtonElement;

    function group(label: string, names: string[] | null) {
        const g = document.createElement("optgroup");
        g.label = label;
        g.append(...(names ?? []).map((name) => new Option(name, name)));
        return g;
    }

    async function loadSounds() {
        const sounds = await SettingsService.Sounds();
        const groups = [group("Custom", sounds.custom), group("System", sounds.system)];
        soundSelect.replaceChildren(...groups.filter((g) => g.children.length > 0));
    }

    function updateRemove() {
        removeButton.disabled = !soundSelect.value.includes(".");
    }

    async function load() {
        const [cfg, login] = await Promise.all([SettingsService.Get(), SettingsService.LaunchAtLogin(), loadSounds()]);
        loginToggle.checked = login;
        for (const name of toggles) {
            checkbox(name).checked = cfg[name];
        }
        soundSelect.value = cfg.soundName;
        trayModeSelect.value = cfg.trayMode;
        bannerPosSelect.value = cfg.bannerPosition;
        bannerPosSelect.disabled = !cfg.banner;
        checkbox("trayFlash").disabled = cfg.trayMode === "never";
        soundSelect.disabled = !cfg.sound;
        updateRemove();
    }

    const loginToggle = form.elements.namedItem("launchAtLogin") as HTMLInputElement;

    form.addEventListener("change", async (event) => {
        if (event.target === loginToggle) {
            try {
                await SettingsService.SetLaunchAtLogin(loginToggle.checked);
                flash(loginToggle.checked ? "Starts at login" : "Won't start at login");
            } catch (err) {
                loginToggle.checked = !loginToggle.checked;
                flash(`Login item failed: ${err}`);
            }
            return;
        }
        updateRemove();
        soundSelect.disabled = !checkbox("sound").checked;
        bannerPosSelect.disabled = !checkbox("banner").checked;
        checkbox("trayFlash").disabled = trayModeSelect.value === "never";
        try {
            await SettingsService.Save(read());
            flash("Saved");
        } catch (err) {
            flash(`Save failed: ${err}`);
        }
    });

    document.getElementById("add-sound")!.addEventListener("click", async () => {
        try {
            const name = await SettingsService.AddSound();
            if (!name) return;
            await loadSounds();
            soundSelect.value = name;
            form.dispatchEvent(new Event("change"));
        } catch (err) {
            flash(`Add failed: ${err}`);
        }
    });

    removeButton.addEventListener("click", async () => {
        const name = soundSelect.value;
        try {
            await SettingsService.RemoveSound(name);
            await loadSounds();
            soundSelect.value = "Glass";
            form.dispatchEvent(new Event("change"));
        } catch (err) {
            flash(`Remove failed: ${err}`);
        }
    });

    document.getElementById("preview")!.addEventListener("click", () => {
        SettingsService.PreviewSound(soundSelect.value).catch((err) => flash(`${err}`));
    });

    document.getElementById("test")!.addEventListener("click", () => {
        SettingsService.Test().then(() => flash("Yippee-ki-yay!"));
    });

    load().catch((err) => flash(`Load failed: ${err}`));
}
