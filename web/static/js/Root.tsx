import React, { useEffect, useState } from "react";
import ReactDOM from "react-dom";
import { BrowserRouter, Switch, Link, Route } from "react-router-dom";
import { ToastContainer } from "react-toastify";
import ScrollToTop from "./components/ScrollToTop";

// @ts-expect-error
import logo from "../images/rich-destiny.png";
import 'react-toastify/dist/ReactToastify.min.css';

import Home from "./Home";
import Download from "./Download";
import FAQ from "./FAQ";
import ControlPanel from "./ControlPanel";

function Root() {
    return <BrowserRouter>
        <ToastContainer autoClose={10000} />

        {/* <InfoBanner /> */}

        <div id="nav">
            <Link to="/">
                <img src={logo} alt="icon" width="40" height="40"/>
                &nbsp;&nbsp;rich-destiny
            </Link>
            <label htmlFor="hamburger">&#9776;</label>
            <input type="checkbox" id="hamburger"/>
            <ul onClick={() => (document.getElementById("hamburger") as HTMLInputElement).checked = false}>
                <li>
                    <Link to="/download">Download</Link>
                </li>
                <li>
                    <Link to="/faq">FAQ</Link>
                </li>
                <li>
                    <Link to="/cp">Control Panel</Link>
                </li>
                <li className="float-right">
                    <a href="https://github.com/lieuweberg/rich-destiny" target="_blank"
                    rel="noopener noreferrer">GitHub</a>
                </li>
                <li className="float-right">
                    <a href="https://discord.gg/UNU4UXp" target="_blank" rel="noopener noreferrer">Discord</a>
                </li>
            </ul>
        </div>

        <div id="view">
            <ScrollToTop />
            <Switch>
                <Route exact path="/">
                    <Home />
                </Route>
                <Route path="/download">
                    <Download />
                </Route>
                <Route path="/faq">
                    <FAQ />
                </Route>
                <Route path="/cp">
                    <ControlPanel />
                </Route>
                <Route path="*">
                    <div className="generic-text">
                        <h1>404</h1>
                        <p>This page does not exist or was removed.</p>
                        <p>Head back <Link to="/">home</Link>.</p>
                    </div>
                </Route>
            </Switch>
        </div>

        <div id="footer">
            <div>
                <p>2020-2021 &copy; <a href="https://lieuweberg.com" target="_blank"
                    rel="noopener noreferrer">lieuwe_berg</a> <br/>
                Destiny 2 and its related assets belong to Bungie, Inc.</p>
            </div>
            <div>
                <a href="https://discord.gg/UNU4UXp" target="_blank"
                    rel="noopener noreferrer">Discord</a>
                <a href="https://github.com/lieuweberg/rich-destiny" target="_blank"
                    rel="noopener noreferrer">GitHub</a>
                <a href="https://twitter.com/richdestinyapp" target="_blank"
                    rel="noopener noreferrer">Twitter</a>
            </div>
        </div>
    </BrowserRouter>
}

function InfoBanner() {
    // The ID of the latest banner. Add 1 to this number when you create a new one! This way it is
    // dismissable across sessions. Yes, it could probably be done better.
    //
    // If you add a new banner, you may have to uncomment the InfoBanner on the actual page above.
    let bannerID = "1";
    let [dismissed, setDismissed] = useState(localStorage.getItem("bannerDismissed"));
    useEffect(() => {}, [dismissed]);

    if (dismissed == bannerID) {
        return null;
    } else {
        return (
            <div id="info-banner" title="Click to dismiss" onClick={() => {
                localStorage.setItem("bannerDismissed", bannerID);
                setDismissed(bannerID);
            }}>
                <p>rich-destiny has moved to its own cozy place on the internet, <code>richdestiny.app</code>!</p>
            </div>
        )
    }
}

const root = document.getElementById("root");
ReactDOM.render(<Root />, root);