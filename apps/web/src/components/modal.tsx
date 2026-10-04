"use client";
import { useEffect, useRef } from "react";
export function Modal({
  children,
  onClose,
  titleId,
  wide = false,
}: {
  children: React.ReactNode;
  onClose: () => void;
  titleId: string;
  wide?: boolean;
}) {
  const ref = useRef<HTMLDialogElement>(null);
  useEffect(() => {
    const el = ref.current;
    el?.showModal();
    return () => el?.close();
  }, []);
  return (
    <dialog
      ref={ref}
      aria-labelledby={titleId}
      onCancel={(event) => {
        event.preventDefault();
        onClose();
      }}
      className={`fixed inset-0 m-auto max-h-[90vh] w-[calc(100%-40px)] ${wide ? "max-w-3xl" : "max-w-lg"} overflow-y-auto rounded-[14px] border-0 bg-white p-0 text-ink shadow-xl backdrop:bg-[#172d33]/40`}
    >
      {children}
    </dialog>
  );
}
