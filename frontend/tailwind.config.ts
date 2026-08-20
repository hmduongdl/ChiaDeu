import type { Config } from 'tailwindcss'

const config: Config = {
  content: [
    './src/pages/**/*.{js,ts,jsx,tsx,mdx}',
    './src/components/**/*.{js,ts,jsx,tsx,mdx}',
    './src/app/**/*.{js,ts,jsx,tsx,mdx}',
  ],
  darkMode: "class",
  theme: {
      extend: {
          "colors": {
              "on-error-container": "#93000a",
              "on-secondary-fixed": "#2a1700",
              "inverse-on-surface": "#ebf1ff",
              "outline-variant": "#bec9c2",
              "surface-bright": "#f9f9ff",
              "on-primary-fixed-variant": "#00513b",
              "on-secondary-fixed-variant": "#653e00",
              "surface-container": "#e7eefe",
              "secondary-fixed-dim": "#ffb95f",
              "error": "#ba1a1a",
              "primary-container": "#065f46",
              "surface-container-low": "#f0f3ff",
              "on-primary-fixed": "#002116",
              "on-error": "#ffffff",
              "surface-variant": "#dce2f3",
              "on-tertiary": "#ffffff",
              "on-tertiary-container": "#ffb4ad",
              "surface-tint": "#1b6b51",
              "outline": "#6f7973",
              "surface-container-highest": "#dce2f3",
              "tertiary-container": "#823f3a",
              "surface": "#f9f9ff",
              "tertiary-fixed": "#ffdad6",
              "on-surface-variant": "#3f4944",
              "inverse-primary": "#8bd6b6",
              "secondary-fixed": "#ffddb8",
              "inverse-surface": "#2a313d",
              "on-background": "#151c27",
              "surface-container-high": "#e2e8f8",
              "tertiary-fixed-dim": "#ffb3ac",
              "on-tertiary-fixed": "#3b0908",
              "on-secondary": "#ffffff",
              "secondary": "#855300",
              "on-tertiary-fixed-variant": "#73332f",
              "on-primary": "#ffffff",
              "primary": "#004532",
              "secondary-container": "#fea619",
              "on-primary-container": "#8bd6b7",
              "background": "#f9f9ff",
              "surface-container-lowest": "#ffffff",
              "on-secondary-container": "#684000",
              "surface-dim": "#d3daea",
              "primary-fixed": "#a6f2d1",
              "tertiary": "#652925",
              "primary-fixed-dim": "#8bd6b6",
              "error-container": "#ffdad6",
              "on-surface": "#151c27"
          },
          "borderRadius": {
              "DEFAULT": "0.25rem",
              "lg": "0.5rem",
              "xl": "0.75rem",
              "2xl": "20px",
              "full": "9999px"
          },
          "spacing": {
              "gutter": "16px",
              "container-padding": "20px",
              "sm": "12px",
              "md": "16px",
              "base": "4px",
              "lg": "24px",
              "xl": "32px",
              "xs": "8px"
          },
          "fontFamily": {
              "sans": ["Inter", "sans-serif"],
              "label-md": ["Inter", "sans-serif"],
              "headline-lg": ["Inter", "sans-serif"],
              "headline-md": ["Inter", "sans-serif"],
              "label-bold": ["Inter", "sans-serif"],
              "headline-lg-mobile": ["Inter", "sans-serif"],
              "body-lg": ["Inter", "sans-serif"],
              "headline-sm": ["Inter", "sans-serif"],
              "body-md": ["Inter", "sans-serif"],
              "body-sm": ["Inter", "sans-serif"]
          },
          "fontSize": {
              "label-md": ["12px", { "lineHeight": "16px", "fontWeight": "500" }],
              "headline-lg": ["30px", { "lineHeight": "38px", "letterSpacing": "-0.02em", "fontWeight": "700" }],
              "headline-md": ["24px", { "lineHeight": "32px", "letterSpacing": "-0.01em", "fontWeight": "600" }],
              "label-bold": ["12px", { "lineHeight": "16px", "fontWeight": "600" }],
              "headline-lg-mobile": ["26px", { "lineHeight": "32px", "fontWeight": "700" }],
              "body-lg": ["18px", { "lineHeight": "28px", "fontWeight": "400" }],
              "headline-sm": ["20px", { "lineHeight": "28px", "fontWeight": "600" }],
              "body-md": ["16px", { "lineHeight": "24px", "fontWeight": "400" }],
              "body-sm": ["14px", { "lineHeight": "20px", "fontWeight": "400" }]
          },
          keyframes: {
            'fade-in-up': {
              '0%': {
                opacity: '0',
                transform: 'translateY(20px)',
              },
              '100%': {
                opacity: '1',
                transform: 'translateY(0)',
              },
            },
          },
          animation: {
            'fade-in-up': 'fade-in-up 0.5s ease-out forwards',
          }
      }
  },
  plugins: [
    require('@tailwindcss/forms'),
    require('@tailwindcss/container-queries')
  ],
}
export default config