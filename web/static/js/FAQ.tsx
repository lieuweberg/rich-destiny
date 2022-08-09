import React from "react";
import { marked } from "marked";

import "../css/faq.scss";

import questions from "../faq-questions.json";

// @ts-expect-error
import decoration from "../images/s17-nightmare1.webp";

export default function() {
    return <>
        <img id="faq-decoration" className="sidebar-decoration" src={decoration} alt="" />
        <div className="generic-text">
            <h1>Frequently Asked Questions</h1>
            <p>This FAQ functions as an FAQ and troubleshooter in one. Find your question,
                follow the steps and you likely will arrive at your solution. If no solution
                works, go to the <a href="https://discord.gg/UNU4UXp" target="_blank"
                rel="noopener noreferrer">support server</a>.
            </p>
        </div>
        <div className="generic-text">
            <div id="contents" className="boxed">
                <h2>Contents</h2>
                <ul id="questions">
                    {questions.map(({q, a, id}, i) => <li key={i}><a href={"#" + id}>{q}</a></li>)}
                </ul>
            </div>
        </div>
        <div className="generic-text">
            {questions.map(({q, a, id}, i) => (
                <section id={id} key={i}>
                    <h2><a href={"#" + id}>#</a> {q}</h2>
                    <p dangerouslySetInnerHTML={{__html: marked(a)}}></p>
                </section>
            ))}
        </div>
    </>
}