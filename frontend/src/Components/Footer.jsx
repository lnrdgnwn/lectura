import { useState } from "react";
// Impor ikon dari react-icons
import {
  FaInstagram,
  FaTiktok,
  FaFacebookF,
  FaTwitter,
  FaYoutube,
  FaPinterest,
} from "react-icons/fa";
// Ikon untuk panah dropdown
import { MdKeyboardArrowDown } from "react-icons/md";

// Data untuk link footer
const footerSections = [
  {
    id: "teams",
    title: "TEAMS",
    links: ["About", "Newsroom", "Brand Guideline"],
  },
  {
    id: "contacts",
    title: "CONTACTS",
    links: [
      "Translators & Editors",
      "Commercial",
      "Audio business",
      "Help & Service",
      "DMCA Notification",
      "Online service",
      "Vulnerability Report",
    ],
  },
  {
    id: "resources",
    title: "RESOURCES",
    links: [
      "Download Apps",
      "Be an Author",
      "Help Center",
      "Privacy Policy",
      "Terms of Service",
      "Affiliate",
    ],
  },
  {
    id: "referrals",
    title: "REFERRALS",
    links: ["Referral Link 1", "Referral Link 2"],
  },
];

// Komponen Ikon Sosial (agar lebih rapi)
const SocialIcon = ({ children, href = "#" }) => (
  <a href={href} className="text-gray-400 hover:text-white transition-colors">
    {children}
  </a>
);

export default function Footer() {
  // State untuk melacak ID section yang terbuka.
  const [openSection, setOpenSection] = useState(null);

  const handleToggle = (sectionId) => {
    setOpenSection((prevOpenSection) =>
      prevOpenSection === sectionId ? null : sectionId
    );
  };

  return (
    <footer className="md:flex items-center justify-center bg-[#3c424b] text-gray-400 font-sans p-6 md:px-16 md:py-12">
      <div className="flex flex-col max-w-5xl w-full md:flex-row md:justify-between">
        <div className="hidden md:flex flex-col gap-4 md:w-1/4">
          <img src="/img/lectura-white.png" alt="" className="w-50 mb-" />
          <div className="flex space-x-4">
            <SocialIcon>
              <FaInstagram size={20} />
            </SocialIcon>
            <SocialIcon>
              <FaTiktok size={20} />
            </SocialIcon>
            <SocialIcon>
              <FaTwitter size={20} />
            </SocialIcon>
            <SocialIcon>
              <FaFacebookF size={20} />
            </SocialIcon>
            <SocialIcon>
              <FaYoutube size={20} />
            </SocialIcon>
            <SocialIcon>
              <FaPinterest size={20} />
            </SocialIcon>
          </div>
          <p className="text-sm">&copy; 2025 Lectura</p>
        </div>

        {/* === BAGIAN KANAN (Links Grid/Accordion) === */}
        <div className="w-full md:w-4/6 md:grid md:grid-cols-4">
          {footerSections.map((section) => (
            <div
              className="border-b border-gray-700 md:border-none"
              key={section.id}
            >
              {/* --- Judul Section --- */}
              <h4
                className="text-white font-bold py-4 flex justify-between items-center cursor-pointer md:cursor-default md:mb-3 md:py-0"
                onClick={() => handleToggle(section.id)}
              >
                <span>{section.title}</span>
                {/* Ikon Panah (hanya tampil di mobile) */}
                <span className="md:hidden">
                  <MdKeyboardArrowDown
                    size={24}
                    className={`transition-transform duration-300 ${
                      openSection === section.id ? "rotate-180" : "rotate-0"
                    }`}
                  />
                </span>
              </h4>

              {/* --- Daftar Link (Accordion di mobile, List di desktop) --- */}
              <ul
                className={`list-none p-0 m-0 overflow-hidden transition-all duration-300 ease-in-out
                  ${openSection === section.id ? "max-h-96 pb-4" : "max-h-0"}
                  md:max-h-none md:pb-0`}
              >
                {section.links.map((link) => (
                  <li key={link} className="mb-2">
                    <a
                      href="#"
                      className="text-sm hover:text-white transition-colors"
                    >
                      {link}
                    </a>
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>
      </div>

      {/* === BAGIAN BAWAH (Mobile Social & Copyright) === */}
      <div className="md:hidden text-center mt-6 pt-6 border-t border-gray-500">
        <div className="flex justify-center space-x-6 mb-6">
          {/* Menggunakan ikon yang sama dengan gambar mobile */}
          <SocialIcon>
            <FaInstagram size={24} />
          </SocialIcon>
          <SocialIcon>
            <FaTiktok size={24} />
          </SocialIcon>
          <SocialIcon>
            <FaFacebookF size={24} />
          </SocialIcon>
          <SocialIcon>
            <FaTwitter size={24} />
          </SocialIcon>
          <SocialIcon>
            <FaYoutube size={24} />
          </SocialIcon>
        </div>
        <p className="text-sm">&copy; 2025 Lectura</p>
      </div>
    </footer>
  );
}
