"use client";

import React, { useEffect, useState } from 'react';
import { motion, useReducedMotion } from 'framer-motion';
import { User } from 'lucide-react';

interface SplitBillAnimationProps {
  /** Tổng số tiền cần chia */
  totalAmount?: number;
  /** Tên nhóm hoặc sự kiện */
  groupName?: string;
  /** Số người chia */
  numPeople?: number;
}

export default function SplitBillAnimation({
  totalAmount = 280000,
  groupName = 'Bữa tối Tối T7',
  numPeople = 4,
}: SplitBillAnimationProps) {
  const shouldReduceMotion = useReducedMotion();
  const [step, setStep] = useState<'total' | 'split'>('total');

  const splitAmount = totalAmount / numPeople;

  const formatMoney = (amount: number) => {
    return new Intl.NumberFormat('vi-VN', { 
      style: 'currency', 
      currency: 'VND',
      maximumFractionDigits: 0
    }).format(amount);
  };

  useEffect(() => {
    if (shouldReduceMotion) return;

    let timeoutId: NodeJS.Timeout;

    const runCycle = () => {
      // Bắt đầu với trạng thái gộp (hiện tổng tiền)
      setStep('total');
      
      // Giữ 0.9s sau đó chuyển sang trạng thái chia
      timeoutId = setTimeout(() => {
        setStep('split');
        
        // Trạng thái chia giữ khoảng 2.6s (tổng cộng 3.5s 1 chu kỳ)
        timeoutId = setTimeout(() => {
          runCycle();
        }, 2600);
      }, 900);
    };

    runCycle();

    return () => clearTimeout(timeoutId);
  }, [shouldReduceMotion]);

  // Fallback tĩnh nếu user bật prefers-reduced-motion
  if (shouldReduceMotion) {
    return (
      <div className="absolute bottom-6 left-6 right-6 z-20 flex justify-center">
        <div className="glass-card rounded-xl p-4 inline-flex items-center gap-3">
          <div className="w-10 h-10 rounded-full bg-primary flex items-center justify-center text-on-primary">
            <User size={20} />
          </div>
          <div>
            <p className="font-label-bold text-label-bold text-on-surface">{groupName}</p>
            <p className="font-body-sm text-body-sm text-primary">Đã chia đều • {numPeople} người</p>
          </div>
        </div>
      </div>
    );
  }

  // Tọa độ 4 hướng bay ra cho 4 người (trên trái, trên phải, dưới trái, dưới phải)
  const directions = [
    { dx: -70, dy: -60 },
    { dx: 70, dy: -60 },
    { dx: -70, dy: 60 },
    { dx: 70, dy: 60 },
  ];

  const containerVariants = {
    total: {},
    split: {
      transition: {
        staggerChildren: 0.1, // So le khi bay ra
      },
    },
  };

  const totalBadgeVariants = {
    total: { scale: 1, opacity: 1 },
    split: { scale: 0.5, opacity: 0 },
  };

  const avatarVariants = {
    total: { scale: 0, opacity: 0, x: 0, y: 0 },
    split: (custom: { dx: number; dy: number }) => ({
      scale: 1,
      opacity: 1,
      x: custom.dx,
      y: custom.dy,
      transition: {
        type: 'spring' as const,
        stiffness: 300,
        damping: 20,
        mass: 0.8,
      },
    }),
  };

  return (
    <div className="absolute inset-0 z-20 flex items-center justify-center pointer-events-none">
      <motion.div
        className="relative flex items-center justify-center mt-16" // mt-16 để căn xuống phần dưới ảnh
        variants={containerVariants}
        initial="total"
        animate={step}
      >
        {/* Badge Tổng Tiền */}
        <motion.div
          variants={totalBadgeVariants}
          transition={{ type: 'spring', stiffness: 300, damping: 25 }}
          className="absolute glass-card rounded-2xl px-6 py-4 flex flex-col items-center justify-center shadow-lg"
        >
          <span className="font-label-bold text-on-surface mb-1">{groupName}</span>
          <span className="font-headline-sm text-primary">{formatMoney(totalAmount)}</span>
        </motion.div>

        {/* Các Avatar bay ra xung quanh */}
        {directions.map((dir, i) => (
          <motion.div
            key={i}
            custom={dir}
            variants={avatarVariants}
            className="absolute flex flex-col items-center justify-center gap-1"
          >
            <div className="w-10 h-10 rounded-full bg-primary-container text-on-primary-container flex items-center justify-center shadow-md">
              <User size={18} />
            </div>
            <div className="glass-card px-2 py-0.5 rounded-full shadow-sm">
              <span className="font-label-bold text-[10px] text-on-surface">
                {formatMoney(splitAmount)}
              </span>
            </div>
          </motion.div>
        ))}
      </motion.div>
    </div>
  );
}
