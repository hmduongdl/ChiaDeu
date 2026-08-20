"use client"
import Link from 'next/link'
import Image from 'next/image'
import SplitBillAnimation from '@/components/landing/SplitBillAnimation'

export default function Home() {
  return (
    <>
      {/* Top App Bar (Landing Page variant) */}
      <header className="w-full top-0 sticky z-50 bg-background/90 backdrop-blur-md">
        <div className="flex items-center justify-between px-container-padding h-16 w-full max-w-md mx-auto">
          <div className="font-headline-md text-headline-md font-bold text-primary flex items-center gap-2">
            <span className="material-symbols-outlined" style={{ fontVariationSettings: "'FILL' 1" }}>pie_chart</span>
            Chia Đều
          </div>
          <Link href="/login" className="font-label-bold text-label-bold text-primary bg-surface-container-low px-4 py-2 rounded-full hover:bg-surface-variant transition-colors">
            Đăng nhập
          </Link>
        </div>
      </header>

      <main className="flex-grow w-full max-w-md mx-auto px-container-padding pb-24">
        {/* Hero Section */}
        <section className="mt-lg mb-xl animate-fade-in-up">
          <div className="relative w-full h-[320px] rounded-2xl overflow-hidden mb-lg shadow-sm">
            <div className="absolute inset-0 bg-gradient-to-t from-black/60 to-transparent z-10"></div>
            {/* Optimized Next Image */}
            <Image 
              alt="Friends enjoying meal" 
              className="w-full h-full object-cover" 
              src="https://lh3.googleusercontent.com/aida-public/AB6AXuCt8LMk2zBtlO5ThYrpmK0u6LioeFqP3mpteZxewfySD218yosTfQgqr5rfEPvABndr9Pl0-Fo3DpVsR6gE-QrG5rp4uWjUbSMxRv_fkNbX4YnM1OcTvqpi4sJH3zdZh6tKjFLDNknElYwKZGDoqnoxqUi87Mxq6KBcIQ7rAHxWE10n30PbG4VRk1nyrko2vflPS271_Q91icLCZn_Bp5ogTGhWAM3WCmXc7lTW_GOE47RxGEw3fKU_Pw"
              fill
              priority
            />
            <SplitBillAnimation />
          </div>

          <div className="space-y-md text-center">
            <h1 className="font-headline-lg-mobile text-headline-lg-mobile text-on-background">
              Chia tiền minh bạch<br/>
              <span className="text-primary">San sẻ yêu thương</span>
            </h1>
            <p className="font-body-md text-body-md text-on-surface-variant max-w-[280px] mx-auto">
              Giải pháp tối ưu để quản lý chi tiêu nhóm, tất toán nợ nần một cách nhẹ nhàng và công bằng.
            </p>
            <Link href="/register" className="w-full h-14 mt-6 bg-gradient-to-r from-primary to-primary-container text-on-primary font-headline-sm text-headline-sm rounded-2xl shadow-[0_8px_16px_rgba(6,95,70,0.2)] hover:opacity-90 active:scale-[0.98] transition-all flex items-center justify-center gap-2">
              Bắt đầu ngay
              <span className="material-symbols-outlined">arrow_forward</span>
            </Link>
          </div>
        </section>

        {/* Core Value Propositions (Bento Layout) */}
        <section className="mb-xl">
          <h2 className="font-headline-sm text-headline-sm text-on-background mb-lg text-center">Tại sao chọn Chia Đều?</h2>
          
          <div className="grid grid-cols-2 gap-md">
            {/* Feature 1: Full width on mobile bento */}
            <div className="col-span-2 bg-surface-container-low p-md rounded-2xl border border-surface-variant flex gap-md items-start">
              <div className="w-12 h-12 rounded-full bg-surface-container flex-shrink-0 flex items-center justify-center text-primary">
                <span className="material-symbols-outlined" style={{ fontVariationSettings: "'FILL' 1" }}>visibility</span>
              </div>
              <div>
                <h3 className="font-headline-sm text-headline-sm text-on-background mb-1">Minh bạch tuyệt đối</h3>
                <p className="font-body-sm text-body-sm text-on-surface-variant">Theo dõi mọi chi phí trong thời gian thực. Không còn tranh cãi về ai nợ ai.</p>
              </div>
            </div>
            
            {/* Feature 2 */}
            <div className="col-span-1 bg-surface-container-lowest p-md rounded-2xl border border-surface-variant shadow-sm flex flex-col gap-sm">
              <div className="w-10 h-10 rounded-full bg-secondary-container/20 flex items-center justify-center text-secondary-container">
                <span className="material-symbols-outlined">account_balance_wallet</span>
              </div>
              <h3 className="font-label-bold text-label-bold text-on-background">Tất toán nhanh chóng</h3>
              <p className="font-label-md text-label-md text-on-surface-variant">Tích hợp ví điện tử phổ biến.</p>
            </div>
            
            {/* Feature 3 */}
            <div className="col-span-1 bg-surface-container-lowest p-md rounded-2xl border border-surface-variant shadow-sm flex flex-col gap-sm">
              <div className="w-10 h-10 rounded-full bg-primary-container/20 flex items-center justify-center text-primary">
                <span className="material-symbols-outlined">auto_awesome</span>
              </div>
              <h3 className="font-label-bold text-label-bold text-on-background">Thông minh &amp; Tự động</h3>
              <p className="font-label-md text-label-md text-on-surface-variant">Tự động tính toán số tiền đóng.</p>
            </div>
          </div>
        </section>

        {/* Social Proof / Footer CTA */}
        <section className="text-center mt-xl pt-lg border-t border-surface-variant">
          <p className="font-body-md text-body-md text-on-surface-variant mb-md">Tham gia cùng +10,000 nhóm bạn đã tin dùng</p>
          <Link href="/login" className="font-label-bold text-label-bold text-primary hover:underline underline-offset-4">
            Đã có tài khoản? Đăng nhập tại đây
          </Link>
        </section>
      </main>
    </>
  )
}