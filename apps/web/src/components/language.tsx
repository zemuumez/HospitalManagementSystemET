"use client";
import { createContext, useContext, useEffect, useState } from "react";
import am from "@/lib/am.json";
type Locale = "en" | "am";
const dictionary: Record<string, string> = am;
const Context = createContext<{
  locale: Locale;
  setLocale: (l: Locale) => void;
  t: (text: string) => string;
}>({ locale: "en", setLocale: () => {}, t: (s) => s });
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
  return (
    <Context.Provider value={{ locale, setLocale: change, t }}>
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
