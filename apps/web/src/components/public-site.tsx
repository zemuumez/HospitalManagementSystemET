"use client";
import Link from "next/link";
import { useEffect, useState } from "react";
import {
  CalendarDays,
  HeartPulse,
  Hospital,
  Menu,
  Stethoscope,
  Users,
  Clock,
  ArrowRight,
} from "lucide-react";
import { LanguageSwitcher, useLanguage } from "./language";
import { defaultFrontSettings, FRONT_KEY } from "@/lib/front-settings";
const doctors = ["Dr. Avery Reed", "Dr. Robin Patel", "Dr. Quinn Parker"];
const services = [
  "General Medicine",
  "Cardiology",
  "Paediatrics",
  "Laboratory",
  "Dental Care",
  "Emergency Care",
];
const titles: Record<string, string> = {
  home: "Home",
  "about-us": "About Us",
  "our-services": "Services",
  doctors: "Doctors",
  appointment: "Appointment",
  "working-hours": "Working Hours",
  testimonials: "Testimonials",
  "contact-us": "Contact Us",
  "privacy-policy": "Privacy Policy",
  "terms-of-service": "Terms & Conditions",
  register: "Sign Up",
};
export function PublicSite({ page = "home" }: { page?: string }) {
  const { t } = useLanguage();
  const [settings, setSettings] = useState(defaultFrontSettings);
  const [mobile, setMobile] = useState(false);
  const [notice, setNotice] = useState("");
  const [doctor, setDoctor] = useState(doctors[0]);
  const [date, setDate] = useState("");
  useEffect(() => {
    const query = new URLSearchParams(window.location.search);
    const chosen = query.get("doctor");
    if (chosen && doctors.includes(chosen)) setDoctor(chosen);
    const chosenDate = query.get("date");
    if (chosenDate && /^\d{4}-\d{2}-\d{2}$/.test(chosenDate))
      setDate(chosenDate);
    try {
      const saved = localStorage.getItem(FRONT_KEY);
      if (saved) setSettings({ ...defaultFrontSettings, ...JSON.parse(saved) });
    } catch {}
  }, []);
  function submit(e: React.FormEvent) {
    e.preventDefault();
    const form = e.target as HTMLFormElement;
    const passwords = form.querySelectorAll<HTMLInputElement>(
      'input[type="password"]',
    );
    if (passwords.length === 2 && passwords[0].value !== passwords[1].value) {
      setNotice("Password confirmation must match the password.");
      return;
    }
    setNotice(t("Preview saved. No hospital data was changed."));
  }
  const doctorsView = (
    <div className="public-cards">
      {doctors.map((name, i) => (
        <article key={name}>
          <div className="doctor-portrait">
            <Stethoscope size={70} />
          </div>
          <h3>{name}</h3>
          <p>{t(services[i])}</p>
          <Link href={`/appointment?doctor=${encodeURIComponent(name)}`}>
            {t("Book an Appointment")} →
          </Link>
        </article>
      ))}
    </div>
  );
  const servicesView = (
    <div className="public-cards">
      {services.map((name, i) => (
        <article key={name}>
          <span className="service-icon">
            {i % 2 ? <Stethoscope /> : <HeartPulse />}
          </span>
          <h3>{t(name)}</h3>
          <Link href="/appointment">{t("Book an Appointment")} →</Link>
        </article>
      ))}
    </div>
  );
  return (
    <div className="public-site">
      <header className="public-header">
        <Link className="public-brand" href="/">
          <Hospital size={36} />
          <strong>ULSHMS</strong>
        </Link>
        <button
          className="public-menu-button"
          aria-label="Open menu"
          onClick={() => setMobile(!mobile)}
        >
          <Menu />
        </button>
        <nav className={mobile ? "open" : ""} aria-label="Public navigation">
          {[
            ["/", "Home"],
            ["/our-services", "Services"],
            ["/doctors", "Doctors"],
            ["/about-us", "About Us"],
          ].map(([href, label]) => (
            <Link
              key={href}
              className={
                (page === "home" ? href === "/" : href === `/${page}`)
                  ? "active"
                  : ""
              }
              href={href}
            >
              {t(label)}
            </Link>
          ))}
          <details>
            <summary>{t("Our Features")}</summary>
            <div>
              {[
                ["appointment", "Appointment"],
                ["working-hours", "Working Hours"],
                ["testimonials", "Testimonials"],
                ["terms-of-service", "Terms & Conditions"],
                ["privacy-policy", "Privacy Policy"],
              ].map(([url, label]) => (
                <Link key={url} href={`/${url}`}>
                  {t(label)}
                </Link>
              ))}
            </div>
          </details>
          <Link href="/contact-us">{t("Contact Us")}</Link>
          <LanguageSwitcher />
          <Link className="primary" href="/login">
            {t("Sign In")}
          </Link>
        </nav>
      </header>
      {page === "home" ? (
        <>
          <section className="public-hero">
            <div className="public-container hero-layout">
              <div>
                <h6>
                  {settings.home_page_experience} {t("Years of Experience")}
                </h6>
                <h1>{t(settings.home_page_title)}</h1>
                <p>{t(settings.home_page_description)}</p>
                <Link className="primary" href="/register">
                  {t("Sign Up")} <ArrowRight size={16} />
                </Link>
              </div>
              <img
                src={settings.home_page_image}
                alt="Hospital care illustration"
              />
            </div>
          </section>
          <section className="public-container">
            <form className="public-booking" action="/appointment">
              <h2>{t("Book an Appointment")}</h2>
              <select
                aria-label="Select Doctor"
                className="field"
                name="doctor"
              >
                <option value="">{t("Select Doctor")}</option>
                {doctors.map((d) => (
                  <option key={d}>{d}</option>
                ))}
              </select>
              <input
                aria-label="Appointment date"
                name="date"
                className="field"
                type="date"
              />
              <button className="primary">{t("Book Now")}</button>
            </form>
            <div className="public-section">
              <div className="section-title">
                <h6>{t("Easy Solutions")}</h6>
                <h2>{t("Four easy steps to better care")}</h2>
              </div>
              <div className="public-steps">
                {[
                  [Stethoscope, "Choose a Doctor"],
                  [CalendarDays, "Select a Date"],
                  [Users, "Confirm Appointment"],
                  [HeartPulse, "Get Care"],
                ].map(([Icon, label], i) => {
                  const I = Icon as typeof HeartPulse;
                  return (
                    <article key={String(label)}>
                      <span>0{i + 1}</span>
                      <I size={35} />
                      <h3>
                        {t(
                          settings[`home_page_step_${i + 1}_title`] ||
                            String(label),
                        )}
                      </h3>
                      {settings[`home_page_step_${i + 1}_description`] && (
                        <p>{settings[`home_page_step_${i + 1}_description`]}</p>
                      )}
                    </article>
                  );
                })}
              </div>
            </div>
            <div className="public-section hero-layout">
              <img
                src={settings.home_page_certified_doctor_image}
                alt="Care team illustration"
              />
              <div>
                <h6>{t("About Us")}</h6>
                <h2>{t(settings.home_page_box_title)}</h2>
                <p>{t(settings.home_page_box_description)}</p>
                <Link className="primary" href="/about-us">
                  {t("About Us")}
                </Link>
              </div>
            </div>
            <div className="public-section">
              <div className="section-title">
                <h2>{t("Our Services")}</h2>
              </div>
              {servicesView}
            </div>
            <div className="public-section">
              <div className="section-title">
                <h2>{t("Our Doctors")}</h2>
              </div>
              {doctorsView}
            </div>
          </section>
        </>
      ) : (
        <>
          <section className="public-page-heading">
            <h1>{t(titles[page])}</h1>
            <p>
              <Link href="/">{t("Home")}</Link> / {t(titles[page])}
            </p>
          </section>
          <main className="public-container public-section">
            {page === "about-us" ? (
              <div className="hero-layout">
                <img
                  src={settings.home_page_certified_doctor_image}
                  alt="Hospital care team"
                />
                <div>
                  <h2>{t(settings.about_us_title || settings.about_title)}</h2>
                  <p>
                    {t(
                      settings.about_us_description ||
                        settings.about_description,
                    )}
                  </p>
                  <Link className="primary" href="/appointment">
                    {t("Book an Appointment")}
                  </Link>
                </div>
              </div>
            ) : page === "our-services" ? (
              servicesView
            ) : page === "doctors" ? (
              doctorsView
            ) : page === "working-hours" ? (
              <div className="public-form">
                <h2>
                  <Clock className="inline mr-3" />
                  {t("Working Hours")}
                </h2>
                {[
                  "Monday",
                  "Tuesday",
                  "Wednesday",
                  "Thursday",
                  "Friday",
                  "Saturday",
                  "Sunday",
                ].map((d) => (
                  <div className="hours-row" key={d}>
                    <span>{t(d)}</span>
                    <strong>
                      {d === "Sunday" ? "Emergency care" : "08:00 – 17:00"}
                    </strong>
                  </div>
                ))}
              </div>
            ) : page === "testimonials" ? (
              <div className="public-cards">
                {["Alex Morgan", "Jamie Wilson", "Taylor Davis"].map((name) => (
                  <article key={name}>
                    <span className="testimonial-quote">“</span>
                    <p>Sample patient testimonial for frontend review.</p>
                    <h3>{name}</h3>
                  </article>
                ))}
              </div>
            ) : page === "terms-of-service" || page === "privacy-policy" ? (
              <article className="public-form">
                <h2>{t(titles[page])}</h2>
                <p>{settings.terms}</p>
                <p className="form-preview-note">
                  Preview content. Hospital-approved policies have not been
                  supplied.
                </p>
              </article>
            ) : (
              <form className="public-form" onSubmit={submit}>
                <h2>
                  {t(
                    page === "appointment"
                      ? settings.appointment_title
                      : titles[page],
                  )}
                </h2>
                <p className="form-preview-note">
                  Frontend preview · This form does not submit real requests or
                  create accounts.
                </p>
                {notice && (
                  <p className="success" role="status">
                    {notice}
                  </p>
                )}
                <div className="legacy-form">
                  <label>
                    <span className="label">{t("First Name")} *</span>
                    <input
                      className="field"
                      required
                      autoComplete="given-name"
                    />
                  </label>
                  <label>
                    <span className="label">{t("Last Name")} *</span>
                    <input
                      className="field"
                      required
                      autoComplete="family-name"
                    />
                  </label>
                  <label>
                    <span className="label">{t("Email")} *</span>
                    <input
                      className="field"
                      required
                      type="email"
                      autoComplete="email"
                    />
                  </label>
                  <label>
                    <span className="label">{t("Phone")} *</span>
                    <input
                      className="field"
                      required
                      type="tel"
                      autoComplete="tel"
                    />
                  </label>
                  {page === "appointment" && (
                    <>
                      <label>
                        <span className="label">{t("Doctor")} *</span>
                        <select
                          className="field"
                          value={doctor}
                          onChange={(e) => setDoctor(e.target.value)}
                        >
                          {doctors.map((d) => (
                            <option key={d}>{d}</option>
                          ))}
                        </select>
                      </label>
                      <label>
                        <span className="label">{t("Date")} *</span>
                        <input
                          className="field"
                          required
                          type="date"
                          value={date}
                          onChange={(e) => setDate(e.target.value)}
                        />
                      </label>
                      <label>
                        <span className="label">Time *</span>
                        <select className="field" required>
                          <option value="">Select time</option>
                          {[
                            "09:00",
                            "09:30",
                            "10:00",
                            "10:30",
                            "14:00",
                            "14:30",
                          ].map((time) => (
                            <option key={time}>{time}</option>
                          ))}
                        </select>
                      </label>
                    </>
                  )}
                  {page === "register" ? (
                    <>
                      <label>
                        <span className="label">{t("Password")} *</span>
                        <input
                          className="field"
                          type="password"
                          minLength={12}
                          required
                          autoComplete="new-password"
                        />
                      </label>
                      <label>
                        <span className="label">
                          {t("Password Confirmation")} *
                        </span>
                        <input
                          className="field"
                          type="password"
                          minLength={12}
                          required
                          autoComplete="new-password"
                        />
                      </label>
                    </>
                  ) : (
                    <label className="form-span">
                      <span className="label">{t("Message")}</span>
                      <textarea className="field" rows={4} />
                    </label>
                  )}
                </div>
                <button className="primary mt-6">
                  {t(
                    page === "appointment"
                      ? "Book Now"
                      : page === "register"
                        ? "Sign Up"
                        : "Send Message",
                  )}
                </button>
              </form>
            )}
          </main>
        </>
      )}
      <footer className="public-footer">
        <div className="public-container">
          <div>
            <h2>ULSHMS</h2>
            <p>{t("Your health, our priority")}</p>
            <small>Frontend preview · Sample content</small>
          </div>
          <div>
            <h3>{t("Our Features")}</h3>
            <Link href="/appointment">{t("Appointment")}</Link>
            <Link href="/working-hours">{t("Working Hours")}</Link>
            <Link href="/doctors">{t("Doctors")}</Link>
          </div>
          <div>
            <h3>{t("Contact Us")}</h3>
            <Link href="/contact-us">{t("Contact Us")}</Link>
            <Link href="/privacy-policy">{t("Privacy Policy")}</Link>
            <Link href="/terms-of-service">{t("Terms & Conditions")}</Link>
          </div>
        </div>
        <p className="public-copyright">© {new Date().getFullYear()} ULSHMS</p>
      </footer>
    </div>
  );
}
