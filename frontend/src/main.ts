import {initBanner} from "./banner";
import {initSettings} from "./settings";

const hash = window.location.hash;
if (hash.startsWith("#banner")) {
    initBanner(new URLSearchParams(hash.slice(hash.indexOf("?") + 1)));
} else {
    initSettings();
}
