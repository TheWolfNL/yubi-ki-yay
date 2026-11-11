import {SettingsService} from "../bindings/github.com/tjeerdritsma/yubi-ki-yay";

function el<K extends keyof HTMLElementTagNameMap>(tag: K, props: Partial<HTMLElementTagNameMap[K]>) {
    return Object.assign(document.createElement(tag), props);
}

export function initBanner(params: URLSearchParams) {
    const label = params.get("label") ?? "";
    const context = params.get("context") ?? "";

    document.body.className = "banner";
    const text = el("div", {className: "text"});
    text.append(
        el("div", {className: "title", textContent: "👆 Touch your YubiKey"}),
        el("div", {className: "context", textContent: context || label}),
    );
    if (context && label) {
        text.append(el("div", {className: "label", textContent: label}));
    }
    const card = el("div", {className: "banner-card"});
    card.append(el("img", {src: "/appicon.png", alt: ""}), text, el("span", {className: "pulse"}));
    document.body.replaceChildren(card);
    card.addEventListener("click", () => SettingsService.DismissBanner());
}
