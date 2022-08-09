import React, { ChangeEvent, FormEvent } from "react";
import axios from "axios";
import { toast } from "react-toastify";
import ReactTooltip from "react-tooltip";
import semverGte from "semver/functions/gte";

import PresenceCard from "./components/PresenceCard";

import "../css/controlPanel.scss";
import useMemoryState from "./MemoryState";

// @ts-expect-error
import decorationLeft from "../images/s17-nightmare2.webp";
// @ts-expect-error
import decorationRight from "../images/s17-nightmare1.webp";

interface Settings {
    orbitText:      string;
    autoUpdate:     boolean;
    prereleases:    boolean;
    joinGameButton: boolean;
    joinOnlySocial: boolean;
}

const defaultSettings: Settings = {
    orbitText: "",
    autoUpdate: true,
    prereleases: false,
    joinGameButton: false,
    joinOnlySocial: false
}

interface ProgramState extends Settings {
    status:         string;
    debug:          string;
    version:        string;
    name:           string;
    presence:       Presence;
}

interface Presence {
    Details:        string;
    State:          string;
    LargeImage:     string;
    // LargeText:      string;
    SmallImage:     string;
    // SmallText:      string;
    Timestamps?:    any;
    Buttons?:       any;
}

const defaultProgramState: ProgramState = {
    status: "Not installed",
    debug: "NA",
    version: "vX.Y.Z",
    name: "Not logged in",
    ...defaultSettings,
    presence: {
        Details: "Not playing...",
        State: "",
        LargeImage: "destinylogo",
        SmallImage: "hunter",
    }
}

export default function() {
    const [data, setData] = useMemoryState("controlPanelData", defaultProgramState) as [ProgramState, Function];
    const [settings, setSettings] = React.useState<Settings>(defaultSettings);

    // Clear the interval when switching to another page. This acts as component unmount.
    React.useEffect(() => {
        (async () => {
            let script = document.createElement("script");
            script.src = "https://platform.twitter.com/widgets.js";
            script.onload = () => {
                // @ts-ignore
                twttr.widgets.load(document.getElementsByClassName("twitter-timeline")[0])
            }
            document.head.appendChild(script);
        })()

        getData(setData, 0);
        let interval = setInterval(() => {
            getData(setData, interval)
        }, 3000)

        return () => clearInterval(interval);
    }, [])

    // Update settings when new data comes in
    React.useEffect(() => {
        let obj = { ...settings };
        Object.keys(obj).forEach(k => obj[k] = data[k]);
        setSettings(obj);
    }, [data.autoUpdate, data.orbitText, data.prereleases, data.joinGameButton, data.joinOnlySocial]);

    function requiresVersion(version: string) {
        if (data.version == "dev" || data.version == "vX.Y.Z") return null;
        else if (semverGte(data.version, version)) return null;
        else return <code>{version} needed</code>;
    }

    function setSetting(k, v) {
        let obj = { ...settings };
        obj[k] = v;
        setSettings(obj);
    }

    return <>
        <div className="transform-flip">
            <img id="cp-decoration-left" className="sidebar-decoration" src={decorationLeft} alt="" />
            <img id="cp-decoration-right" className="sidebar-decoration" src={decorationRight} alt="" />
        </div>

        <div id="cp" className="generic-text">
            <div className="boxed">
                <div>
                    <h1>Control Panel</h1>
                    <p>Status: {data.status}<br/>
                    Logged in as: {data.name}<br/>
                    Debug: {data.debug}<br/>
                    Version: {data.version}</p>
                </div>
                <div style={{marginLeft: "auto"}}>
                    <h4>Current presence preview</h4>
                    <PresenceCard description={data.presence.Details} state={data.presence.State}
                        largeImage={data.presence.LargeImage} smallImage={data.presence.SmallImage}
                        initialTime={data.presence.Timestamps ? data.presence.Timestamps.Start : null}/>
                </div>
                <div>
                    <h4>Orbit presence preview</h4>
                    <PresenceCard description="In Orbit" state={settings.orbitText} largeImage="destinylogo"/>
                </div>
            </div>

            <div className="boxed">
                <h2>Settings</h2>
                <form>
                    <h4>General</h4>
                    <div>
                        <label>
                            Orbit state text: <input type="text" id="orbitText" placeholder="empty up here..."
                                value={settings.orbitText} onChange={e => setSetting("orbitText", e.target.value)} />
                            &nbsp; <span data-tip="Text to display on the second line of the presence. See
                            the Orbit presence preview to the left. Leave empty to disable.">&#x1f6c8;</span>
                        </label> <br/>

                        <CheckboxInput name="Auto update" json="autoUpdate" value={settings.autoUpdate}
                            update={setSetting} text="Whether to update to the latest releases of
                            rich-destiny automatically. If unticked, you can use the Update button below." />

                        <CheckboxInput name="Prereleases ⚠️" json="prereleases" value={settings.prereleases}
                            update={setSetting} text="Enables prereleases. This option is ⚠️IRREVERSIBLE⚠️. You
                            are fairly expected to report any errors in the support server, however that is of
                            course optional. This option will grant access to early releases that include new
                            features that may possibly not work well. Turning off prereleases will update to a
                            stable NEWER release, and not downgrade." /> {requiresVersion("v0.2.5-1")}
                    </div>

                    <h4>Launch Game button</h4>
                    <div>
                        <CheckboxInput name="Enabled" json="joinGameButton" value={settings.joinGameButton}
                            update={setSetting} text="Adds a 'Launch Game' button to your status that
                            allows anyone (including people without rich-destiny) to launch Destiny 2, simply
                            by clicking it." /> {requiresVersion("v0.2.1")}

                        <CheckboxInput name="Orbit or social spaces only" json="joinOnlySocial"
                            value={settings.joinOnlySocial} update={setSetting} text="When ticked, the Launch
                            Game button will appear only when you're in orbit or social spaces like the Tower.
                            This feature is still here for backwards compatibility, but isn't really necessary
                            since the Join Game button was removed." /> {requiresVersion("v0.1.9")}
                    </div>
                    <a href="#" className="button" onClick={e => {handleFormSubmit(e, settings)}}>Save Settings</a>
                </form>
            </div>

            <div className="boxed">
                <h2>Actions</h2>
                <div id="actions">
                    <a href="http://localhost:35893/login" className="button" target="_blank"
                        rel="noopener noreferrer" data-tip="In case the refresh token has expired, or
                        you want to log in with a different account.">Login with Bungie</a>

                    <a onClick={e => {
                        e.preventDefault();
                        document.getElementById("reconnect").innerHTML = "Reconnecting...";
                        doSimpleGetRequest("http://localhost:35893/action?a=reconnect", 0, () => {
                            document.getElementById("reconnect").innerHTML = "Reconnect to Discord";
                        });
                    }} href="#" className="button" id="reconnect" data-tip="Reconnect to Discord. This is only
                        supposed to be used when this site says you're playing the game, but Discord
                        isn't.">Reconnect to Discord</a> {requiresVersion("v0.2.1")}

                    <a onClick={e => {
                        e.preventDefault();
                        document.getElementById("update").innerHTML = "Updating...";
                        doSimpleGetRequest("http://localhost:35893/action?a=update", 0, () => {
                            document.getElementById("update").innerHTML = "Update";
                        });
                    }} href="#" className="button" id="update" data-tip="Force finding and downloading of
                        the latest version of the program. If it's newer, it's downloaded, but the program
                        has to be restarted for the update to apply.">Update</a>

                    <a onClick={e => {
                        e.preventDefault();
                        document.getElementById("uninstall").innerHTML = "Uninstalling... :(";
                        doSimpleGetRequest("http://localhost:35893/action?a=uninstall", 0, () => {
                            document.getElementById("uninstall").innerHTML = "Uninstalled :(";
                        });
                    }} href="#" className="button" id="uninstall" data-tip="Uninstall rich-destiny from the service
                        manager. Files need to be removed manually!">Uninstall</a> {requiresVersion("v0.2.1")}
                </div>
            </div>

            <div className="boxed">
                <h2>Guardian,</h2>
                <p>We will not cower in fear of Nightmares. We will rise to meet the enemy and confront our darkest fears. I believe in you :)</p>
            </div>

            <div className="boxed">
                <h2>Come hang out</h2>
                <p>Got feedback, questions or interesting ideas? Need a place to vent out about teleporting Overload
                    Captains, or just be with some friendly people? Or come for the opt-in pings when
                    there's a new release.<br/>Come join the Discord server!</p>
                <a href="https://discord.gg/UNU4UXp" target="_blank" rel="noopener noreferrer">
                    <img alt="Discord" src="https://img.shields.io/discord/604679605630009384
                        ?label=Discord&color=6c82cf"/>
                </a>

                <h3>Or show some support</h3>
                <p>By leaving a star on GitHub :)</p>
                <a href="https://github.com/lieuweberg/rich-destiny" target="_blank" rel="noopener noreferrer">
                    <img alt="GitHub stars" src="https://img.shields.io/github/stars/lieuweberg/rich-destiny
                        ?label=GitHub%20stars&color=6c82cf"></img>
                </a>
            </div>

            <div className="boxed">
                <h2>Tweets</h2>
                <div>
                    <a className="twitter-timeline" data-lang="en" data-dnt="true" data-theme="dark"
                        data-chrome="noscrollbar nofooter noheader transparent"
                        href="https://twitter.com/richdestinyapp">Loading @richdestinyapp tweets...</a>
                </div>
            </div>
        </div>
        <ReactTooltip effect="solid" backgroundColor="#18191C"/>
    </>
}

function CheckboxInput({name, json, value, update, text}) {
    return <>
        <label>
            {name}: <input type="checkbox" id={json} checked={value}
            onChange={e => update(json, e.target.checked)} /> &nbsp;
            <span data-tip={text}>&#x1f6c8;</span>
        </label> <br/>
    </>
}

function handleHTTPError(err) {
    if (err.response) {
        toast.error(<p>{err.response.data}</p>);
    } else if (err.request) {
        toast.error(<p>rich-destiny could not be reached. Do you have it installed? Is it running?</p>);
    } else {
        toast.error(<p>Could not make request... <code>${err.message}</code></p>);
    }
}

function getData(setData: Function, interval: number) {
    axios.get("http://localhost:35893/action?a=current", {
        timeout: 1000
    }).then(res => {
        for (let key of Object.keys(res.data.presence)) {
            if (res.data.presence[key] == "" || res.data.presence[key] == null) {
                res.data.presence[key] = defaultProgramState.presence[key];
            }
        }
        setData({ ...defaultProgramState, ...res.data });
    }).catch(err => {
        if (interval != 0) {
            handleHTTPError(err)
            clearInterval(interval)
        }
    })
}

function handleFormSubmit(e: React.MouseEvent<HTMLAnchorElement, MouseEvent>, settings: Settings) {
    e.preventDefault();
    axios.post("http://localhost:35893/action?a=save", { ...settings }, { timeout: 1000 })
    .then(res => {
        toast.dark("Settings saved!")
    }).catch(err => handleHTTPError(err));
}

function doSimpleGetRequest(url: string, timeout: number, callback?: Function) {
    axios.get(url, {
        timeout
    }).then(res => {
        toast.dark(res.data);
    }).catch(err => handleHTTPError(err))
    .then(() => {
        if (callback) callback();
    })
}