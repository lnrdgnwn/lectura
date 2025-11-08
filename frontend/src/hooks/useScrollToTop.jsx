import { useEffect } from "react";
import { useLocation } from "react-router-dom";

export default function useScrollToTop(smooth = true) {
  const { pathname } = useLocation();

  useEffect(() => {
    window.scrollTo({
      top: 0,
    });
  }, [pathname, smooth]);
}
