"use client";
import { createContext, useContext, useEffect, useState } from "react";
import am from "@/lib/am.json";
type Locale = "en" | "am";
const dictionary: Record<string, string> = am;
const Context = createContext<{
  locale: Locale;
  setLocale: (l: Locale) => void;
  t: (text: string) => string;
  formatValue: (text: string) => string;
}>({ locale: "en", setLocale: () => {}, t: (s) => s, formatValue: (s) => s });
export function LanguageProvider({ children }: { children: React.ReactNode }) {
  const [locale, setLocale] = useState<Locale>("en");
  useEffect(() => {
    const saved = localStorage.getItem("hms-language");
    if (saved === "am" || saved === "en") setLocale(saved);
  }, []);
  useEffect(() => {
    document.documentElement.lang = locale;
  }, [locale]);
  function change(l: Locale) {
    setLocale(l);
    localStorage.setItem("hms-language", l);
  }
  const t = (text: string) =>
    locale === "en"
      ? text
      : dictionary[text] || dictionary[text.trim()] || text;
  function formatValue(text: string) {
    const language = locale === "am" ? "am-ET-u-ca-gregory" : "en-GB";
    if (/^\d{4}-\d{2}-\d{2}$/.test(text) && !Number.isNaN(Date.parse(text)))
      return new Intl.DateTimeFormat(language, {
        dateStyle: "medium",
        timeZone: "UTC",
      }).format(new Date(text));
    if (/^-?\d+\.\d{2}$/.test(text))
      return new Intl.NumberFormat(language, {
        minimumFractionDigits: 2,
        maximumFractionDigits: 2,
      }).format(Number(text));
    return t(text);
  }
  useEffect(() => {
    const controls = (
      el: EventTarget | null,
    ): el is HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement =>
      el instanceof HTMLInputElement ||
      el instanceof HTMLSelectElement ||
      el instanceof HTMLTextAreaElement;
    const invalid = (event: Event) => {
      const el = event.target;
      if (!controls(el)) return;
      el.setCustomValidity("");
      if (locale !== "am") return;
      const v = el.validity;
      const key = v.valueMissing
        ? "Please fill out this field."
        : v.typeMismatch
          ? "Please enter a valid email address."
          : v.patternMismatch
            ? "Please match the requested format."
            : v.rangeUnderflow || v.rangeOverflow
              ? "Please use a value within the allowed range."
              : "Please enter a valid value.";
      el.setCustomValidity(dictionary[key]);
    };
    const clear = (event: Event) => {
      if (controls(event.target)) event.target.setCustomValidity("");
    };
    document
      .querySelectorAll<
        HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement
      >("input,select,textarea")
      .forEach((el) => el.setCustomValidity(""));
    document.addEventListener("invalid", invalid, true);
    document.addEventListener("input", clear, true);
    document.addEventListener("change", clear, true);
    return () => {
      document.removeEventListener("invalid", invalid, true);
      document.removeEventListener("input", clear, true);
      document.removeEventListener("change", clear, true);
    };
  }, [locale]);
  return (
    <Context.Provider value={{ locale, setLocale: change, t, formatValue }}>
      {children}
    </Context.Provider>
  );
}
export const useLanguage = () => useContext(Context);
export function LanguageSwitcher() {
  const { locale, setLocale } = useLanguage();
  return (
    <select
      aria-label="Language / ቋንቋ"
      className="language-switcher"
      value={locale}
      onChange={(e) => setLocale(e.target.value as Locale)}
    >
      <option value="en">English</option>
      <option value="am">አማርኛ</option>
    </select>
  );
}
