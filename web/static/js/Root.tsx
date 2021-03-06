import React from "react";
import ReactDOM from "react-dom";
import { BrowserRouter, Switch, Link, Route } from "react-router-dom";
import { ToastContainer } from "react-toastify";
import ScrollToTop from "./components/ScrollToTop";

// @ts-expect-error
import logo from "../images/rich-destiny.png";
import 'react-toastify/dist/ReactToastify.min.css';

import Home from "./Home";
import Download from "./Download";
import ControlPanel from "./ControlPanel";

function Root() {
    return <BrowserRouter basename="/rich-destiny">
        <ToastContainer autoClose={10000}/>

        <div id="nav">
            <Link to="/" id="nav-logo">
                <img src={logo} alt="icon" width="40"/>
                &nbsp;&nbsp;rich-destiny
            </Link>
            <label htmlFor="hamburger">&#9776;</label>
            <input type="checkbox" id="hamburger"/>
            <ul id="nav-items" onClick={() =>
                (document.getElementById("hamburger") as HTMLInputElement).checked = false}>
                <li className="nav-item">
                    <Link to="/download">Download</Link>
                </li>
                <li className="nav-item">
                    <Link to="/cp">Control Panel</Link>
                </li>
                <li className="nav-item float-right">
                    <a href="https://github.com/lieuweberg/rich-destiny">GitHub</a>
                </li>
                <li className="nav-item float-right">
                    <a href="https://discord.gg/UNU4UXp">Discord</a>
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
                <Route path="/cp">
                    <ControlPanel />
                </Route>
                <Route path="*">
                    <div className="generic-text top-text">
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
                <a href="https://discord.gg/UNU4UXp">Discord</a>
                <a href="https://github.com/lieuweberg/rich-destiny">GitHub</a>
            </div>
        </div>
    </BrowserRouter>
}

const root = document.getElementById("root");
ReactDOM.render(<Root />, root);