import { useState, useRef, useEffect } from "react";
import { Swiper, SwiperSlide } from "swiper/react";
import { Pagination, Navigation, Autoplay } from "swiper/modules";
import { FastAverageColor } from "fast-average-color";

export default function Slider() {
  const slides = [
    { id: 1, image: "/img/banner1.jpg", alt: "Banner 1" },
    { id: 2, image: "/img/banner2.jpg", alt: "Banner 2" },
    { id: 3, image: "/img/banner3.jpg", alt: "Banner 3" },
  ];

  const [bgColors, setBgColors] = useState({});
  const imgRefs = useRef({});
  const fac = useRef(new FastAverageColor()).current;

  const handleImageLoad = (id) => {
    const img = imgRefs.current[id];
    if (!img) return;
    try {
      const color = fac.getColor(img);
      setBgColors((prev) => ({ ...prev, [id]: color.rgb }));
    } catch (err) {
      console.warn("Color extraction failed:", err);
    }
  };

  useEffect(() => {
    return () => fac.destroy();
  }, [fac]);

  return (
    <section className="w-full">
      <Swiper
        slidesPerView={1}
        loop
        pagination={{ clickable: true }}
        navigation
        autoplay={{
          delay: 4000,
          disableOnInteraction: false,
        }}
        modules={[Pagination, Navigation, Autoplay]}
        className="mySwiper"
      >
        {slides.map((slide) => {
          const bgColor = bgColors[slide.id] || "#1a1a1a";
          return (
            <SwiperSlide key={slide.id}>
              <div
                className="
                  relative flex items-center justify-center overflow-hidden
                  h-[200px] sm:h-[250px] md:h-[300px] lg:h-[400px] xl:h-[500px]
                "
                style={{
                  background: `${bgColor}`,
                  transition: "background 0.6s ease",
                }}
              >
                <img
                  ref={(el) => (imgRefs.current[slide.id] = el)}
                  src={slide.image}
                  alt={slide.alt}
                  crossOrigin="anonymous"
                  onLoad={() => handleImageLoad(slide.id)}
                  className="absolute inset-0 w-full h-full object-cover sm:object-contain object-center"
                />
              </div>
            </SwiperSlide>
          );
        })}
      </Swiper>
    </section>
  );
}
