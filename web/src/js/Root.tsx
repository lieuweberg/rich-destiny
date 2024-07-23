import React, { useEffect, useState } from "react";
import ReactDOM from "react-dom";
import { BrowserRouter, Switch, Link, Route } from "react-router-dom";
import { ToastContainer } from "react-toastify";
import ScrollToTop from "./components/ScrollToTop";

// @ts-expect-error
import logo from "../../static/images/rich-destiny.png";
// @ts-expect-error
import footerWaves from "../../static/images/footer-waves.svg";

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
                <img src={logo} alt="icon" width="40" height="40" />
                &nbsp;&nbsp;rich-destiny
            </Link>
            <label htmlFor="hamburger">&#9776;</label>
            <input type="checkbox" id="hamburger" />
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
                    <a href="https://richdestiny.app/discord" target="_blank" rel="noopener noreferrer">Discord</a>
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
                <Route path="/auth-expired">
                    <div className="generic-text">
                        <h1>Authentication expired</h1>
                        <p>Your authentication details (refresh token) for rich-destiny have expired. Click the button below to
                            log in with Bungie again.</p>
                        <a href="http://localhost:35893/login" className="button" rel="noopener noreferrer">Login with Bungie</a>
                    </div>
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
            <img className="svg-spacer" src={footerWaves} alt="" />
            <div>
                <div>
                    <p>2020-2023 &copy; <a href="https://lieuweberg.com" target="_blank"
                        rel="noopener noreferrer">lieuwe_berg</a> <br />
                        Destiny 2 and its related assets belong to Bungie, Inc.</p>
                </div>
                <div>
                    <a href="https://richdestiny.app/discord" target="_blank"
                        rel="noopener noreferrer">Discord</a>
                    <a href="https://github.com/lieuweberg/rich-destiny" target="_blank"
                        rel="noopener noreferrer">GitHub</a>
                    <a href="https://twitter.com/richdestinyapp" target="_blank"
                        rel="noopener noreferrer">Twitter</a>
                </div>
            </div>
        </div>
    </BrowserRouter>
}

function InfoBanner() {
    // The ID of the latest banner. Add 1 to this number when you create a new one! This way it is
    // dismissable across sessions. Yes, it could probably be done better.
    //
    // If you add a new banner, you may have to uncomment the InfoBanner on the actual page above.
    let bannerID = "2";
    let [dismissed, setDismissed] = useState(localStorage.getItem("bannerDismissed"));
    useEffect(() => { }, [dismissed]);

    if (dismissed == bannerID) {
        return null;
    } else {
        return (
            <a href="https://predict.wastedondestiny.com/" target="_blank">
                <div id="info-banner" /*title="Click to dismiss" onClick={() => {
                    localStorage.setItem("bannerDismissed", bannerID);
                    setDismissed(bannerID);
                }}*/>
                    <p>DESTINY 2: THE FINAL SHAPE PREDICTIONS - Special occasion: <i>The Final Shape</i> is almost out! Click here to try and score points by predicting what's to come!</p>
                </div>
            </a>
        )
    }
}

const root = document.getElementById("root");
ReactDOM.render(<Root />, root);