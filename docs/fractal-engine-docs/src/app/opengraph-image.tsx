import { ImageResponse } from 'next/og';

export const size = {
  width: 1200,
  height: 630,
};

export const contentType = 'image/png';

export default async function Image() {
  return new ImageResponse(
    (
      <div
        style={{
          width: '100%',
          height: '100%',
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          justifyContent: 'center',
          background: 'linear-gradient(135deg, #0f172a 0%, #1f2937 100%)',
          color: 'white',
          fontSize: 72,
          fontWeight: 700,
          letterSpacing: '-0.02em',
        }}
      >
        <div style={{ fontSize: 40, opacity: 0.8, marginBottom: 16 }}>Dogecoin Foundation</div>
        <div>Fractal Engine Docs</div>
      </div>
    ),
    size,
  );
}


