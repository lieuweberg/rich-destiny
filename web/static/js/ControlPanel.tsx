import React, { ChangeEvent, FormEvent } from "react";
import axios from "axios";
import { toast } from "react-toastify";
import ReactTooltip from "react-tooltip";
import semverGte from "semver/functions/gte";

import PresenceCard from "./components/PresenceCard";

import "../css/controlPanel.scss";
import useMemoryState from "./MemoryState";

interface APIResponse {
    status:         string;
    debug:          string;
    version:        string;
    name:           string;
    orbitText:      string;
    autoUpdate:     boolean;
    joinGameCode:   string;
    joinOnlySocial: boolean;
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

const DefaultData: APIResponse = {
    status: "Not installed",
    debug: "NA",
    version: "vX.Y.Z",
    name: "Not logged in",
    orbitText: "",
    autoUpdate: true,
    joinGameCode: "",
    joinOnlySocial: false,
    presence: {
        Details: "Not playing...",
        State: "",
        LargeImage: "destinylogo",
        SmallImage: "hunter",
    }
}

export default function() {
    const [data, setData] = useMemoryState("controlPanelData", DefaultData) as [APIResponse, Function];
    const [intervalID, setIntervalID] = React.useState(-1);
    const [orbitTextValue, setOrbitTextValue] = React.useState("");
    const [autoUpdateValue, setAutoUpdateValue] = React.useState(false);
    const [joinGameCodeValue, setJoinGameCodeValue] = React.useState("");
    const [joinOnlySocialValue, setJoinOnlySocialValue] = React.useState(false);

    if (intervalID == -1) {
        let interval = setInterval(() => {
            getData(setData, interval)
        }, 3000)
        getData(setData, interval);
        setIntervalID(interval);
    }

    // Clear the interval when switching to another page. [intervalID] makes it so this only happens
    // when intervalID changes, and that is once, so stonks. This acts as component unmount.
    React.useEffect(() => {
        return () => {
            clearInterval(intervalID);
        }
    }, [intervalID])
    React.useEffect(() => {
        // update settings when new data comes in
        setOrbitTextValue(data.orbitText);
        setAutoUpdateValue(data.autoUpdate);
        setJoinGameCodeValue(data.joinGameCode);
        setJoinOnlySocialValue(data.joinOnlySocial)
    }, [data.orbitText, data.joinGameCode, data.autoUpdate, data.joinOnlySocial])

    function requiresVersion(version: string) {
        if (data.version == "dev" || data.version == "vX.Y.Z") return null;
        else if (semverGte(data.version, version)) return null;
        else return <code>{version} needed! (current: {data.version})</code>;
    }

    return <> <div id="cp" className="generic-text top-text">
        <div>
            <h1>Control Panel</h1>
            <p>Status: {data.status}<br/>
            Logged in as: {data.name}<br/>
            Debug: {data.debug}<br/>
            Version: {data.version}</p>
        </div>
        <div>
            <h4>Current presence preview</h4>
            <PresenceCard description={data.presence.Details} state={data.presence.State}
                largeImage={data.presence.LargeImage} smallImage={data.presence.SmallImage}
                initialTime={data.presence.Timestamps ? data.presence.Timestamps.Start : null}/>
        </div>
        <div>
            <h2>Settings</h2>
            <form>
                <h4>General</h4>
                <div>
                    <label>
                        Orbit state text: <input type="text" id="orbitText" placeholder="empty up here..."
                            value={orbitTextValue} onChange={e => setOrbitTextValue(e.target.value)} />
                        &nbsp; <span data-tip="Text to display on the second line of the presence. See
                        the preview to the right. Leave empty to disable.">&#x1f6c8;</span>
                    </label> <br/>
                    <label>
                        Auto update: <input type="checkbox" id="autoUpdate" checked={autoUpdateValue}
                            onChange={e => setAutoUpdateValue(e.target.checked)} />
                        &nbsp; <span data-tip="Whether to update to the latest releases of rich-destiny
                        automatically. If unticked, you can use the Update button below.">&#x1f6c8;</span>
                    </label> <br/>
                </div>

                <h4>Join Game button</h4>
                <div>
                    <label>
                        Code: <input type="text" id="joinGameCode" placeholder="76561198237606311"
                            value={joinGameCodeValue} onChange={e => setJoinGameCodeValue(e.target.value)} />
                        &nbsp; <span data-tip="Adds a 'Join Game' button to your status that allows anyone
                        (including people without rich-destiny) to join your fireteam, simply by clicking it.
                        <br/><br/> Enter the number from <code>/id</code> here (your SteamID64). Leave empty
                        to disable." data-html>&#x1f6c8;</span> {requiresVersion("v0.1.8")}
                    </label> <br/>
                    <label>
                        Orbit or social spaces only: <input type="checkbox" id="joinOnlySocial"
                        checked={joinOnlySocialValue}
                            onChange={e => setJoinOnlySocialValue(e.target.checked)} />
                        &nbsp; <span data-tip="When ticked, the Join Game button will appear only when you're
                        in orbit or social spaces like the Tower, preventing people from joining mid-game
                        or when you're trying to complete an activity solo, like a dungeon.">&#x1f6c8;
                        </span> {requiresVersion("v0.1.9")}
                    </label> <br/>
                </div>
                <a href="#" className="button" onClick={handleFormSubmit}>Save Settings</a>
            </form>
        </div>
        <div>
            <h4>Orbit presence preview</h4>
            <PresenceCard description="In orbit" state={orbitTextValue} largeImage="destinylogo"/>
        </div>
        <div>
            <h2>Actions</h2>
            <a href="http://localhost:35893/login" className="button" target="_blank"
                rel="noopener noreferrer" data-tip="In case the refresh token has expired, or
                you want to log in with a different account.">Reauthenticate</a> <br/>
            <a onClick={() => {
                document.getElementById("update").innerHTML = "Updating...";
                doSimpleGetRequest("http://localhost:35893/action?a=update", 0, () => {
                    document.getElementById("update").innerHTML = "Update";
                });
            }} href="#" className="button" id="update" data-tip="Force finding and installing of
            the latest version of the program. If the version is the same, nothing happens. If it's
            newer, it's installed, but the program has to be restarted for an update to apply.">Update</a>
        </div>
    </div> <ReactTooltip effect="solid" backgroundColor="#18191C"/> </>
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
                res.data.presence[key] = DefaultData.presence[key];
            }
        }
        setData({ ...DefaultData, ...res.data });
    }).catch(err => {
        handleHTTPError(err)
        clearInterval(interval)
    })
}

function handleFormSubmit() {
    axios.post("http://localhost:35893/action?a=save", {
        orbitText: (document.getElementById("orbitText") as HTMLInputElement).value,
        autoUpdate: (document.getElementById("autoUpdate") as HTMLInputElement).checked,
        joinGameCode: (document.getElementById("joinGameCode") as HTMLInputElement).value,
        joinOnlySocial: (document.getElementById("joinOnlySocial") as HTMLInputElement).checked
    }, { timeout: 1000 })
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