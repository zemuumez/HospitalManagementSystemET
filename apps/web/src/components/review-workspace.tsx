"use client";

import { useLanguage } from "@/components/language";
import { useState } from "react";
import Link from "next/link";
import { Search, Eye, Star, X } from "lucide-react";

export type ReviewWorkspaceProps = {
  id?: string;
};

interface ReviewItem {
  id: string;
  patientName: string;
  patientEmail: string;
  patientInitials: string;
  doctorName: string;
  doctorEmail: string;
  doctorDepartment: string;
  rating: number;
  status: "Approved" | "Pending" | "Rejected";
  comment: string;
  createdAt: string;
}

export function ReviewWorkspace({ id = "review" }: ReviewWorkspaceProps) {
  const { t } = useLanguage();

  const [search, setSearch] = useState("");
  const [pageSize, setPageSize] = useState(10);
  const [page, setPage] = useState(1);

  // Selected review for details modal (Screenshot 180643)
  const [selectedReview, setSelectedReview] = useState<ReviewItem | null>(null);

  // Reviews data matching Screenshot 180553
  const [reviews, setReviews] = useState<ReviewItem[]>([
    {
      id: "REV-1",
      patientName: "Trith Shah",
      patientEmail: "tirth@gmail.com",
      patientInitials: "TS",
      doctorName: "Harish Mohan",
      doctorEmail: "vatsal@gmail.com",
      doctorDepartment: "Kidney",
      rating: 5,
      status: "Approved",
      comment:
        "Dr. Harish Mohan was very professional and explained everything clearly. The treatment was excellent, and I felt comfortable throughout the consultation. Highly recommended.",
      createdAt: "04 Aug 2026, 06:15 AM",
    },
    {
      id: "REV-2",
      patientName: "SAN K",
      patientEmail: "mwpsquade@gmail.com",
      patientInitials: "SK",
      doctorName: "Dharman K",
      doctorEmail: "smartentry8@gmail.com",
      doctorDepartment: "General Medicine",
      rating: 5,
      status: "Approved",
      comment:
        "Prompt care, clear diagnosis, and compassionate staff. Great experience throughout my admission.",
      createdAt: "12 Aug 2026, 10:20 AM",
    },
    {
      id: "REV-3",
      patientName: "Abebe Kebede",
      patientEmail: "abebe@example.com",
      patientInitials: "AK",
      doctorName: "Ahmed Doctor",
      doctorEmail: "doctorahmed@gmail.com",
      doctorDepartment: "Cardiology",
      rating: 4,
      status: "Approved",
      comment:
        "Very thorough cardiac checkup and follow-up plan. The hospital facilities are modern and spotless.",
      createdAt: "22 Sep 2026, 02:45 PM",
    },
  ]);

  return (
    <div className="legacy-workspace">
      {/* Top subtabs */}
      <div className="module-subtabs-nav">
        <Link href="/modules/review" className="module-subtab-link active">
          {t("Review")}
        </Link>
      </div>

      <div className="billing-card">
        <div className="billing-toolbar">
          <div className="billing-search-box">
            <Search size={16} />
            <input
              type="text"
              placeholder={t("Search")}
              value={search}
              onChange={(e) => setSearch(e.target.value)}
            />
          </div>
        </div>

        <div className="table-responsive">
          <table className="billing-table w-100">
            <thead>
              <tr>
                <th>{t("REVIEWED BY")} ↕</th>
                <th>{t("DOCTORS")} ↕</th>
                <th>{t("STATUS")} ↕</th>
                <th>{t("COMMENT")}</th>
                <th>{t("ACTION")}</th>
              </tr>
            </thead>
            <tbody>
              {reviews
                .filter((r) =>
                  (r.patientName + " " + r.doctorName + " " + r.comment)
                    .toLowerCase()
                    .includes(search.toLowerCase()),
                )
                .map((rev) => (
                  <tr key={rev.id}>
                    <td>
                      <div className="d-flex align-items-center gap-2">
                        <div className="patient-avatar-circle">
                          {rev.patientInitials}
                        </div>
                        <div>
                          <div className="fw-semibold text-primary">
                            {rev.patientName}
                          </div>
                          <div className="text-secondary small">
                            {rev.patientEmail}
                          </div>
                        </div>
                      </div>
                    </td>
                    <td>
                      <div className="d-flex align-items-center gap-2">
                        <div className="doctor-avatar-circle">
                          {rev.doctorName.slice(0, 2).toUpperCase()}
                        </div>
                        <div>
                          <div className="d-flex align-items-center gap-2">
                            <span className="fw-semibold text-primary">
                              {rev.doctorName}
                            </span>
                            <div className="d-flex align-items-center text-warning">
                              {[...Array(rev.rating)].map((_, i) => (
                                <Star
                                  key={i}
                                  size={13}
                                  fill="#f59e0b"
                                  color="#f59e0b"
                                />
                              ))}
                            </div>
                          </div>
                          <div className="text-secondary small">
                            {rev.doctorEmail}
                          </div>
                        </div>
                      </div>
                    </td>
                    <td>
                      <span className="badge-green">{rev.status}</span>
                    </td>
                    <td style={{ maxWidth: "450px" }}>
                      <p className="m-0 text-secondary small text-truncate-2">
                        {rev.comment}
                      </p>
                    </td>
                    <td>
                      <button
                        className="btn-icon-blue-link"
                        title={t("View Details")}
                        onClick={() => setSelectedReview(rev)}
                      >
                        <Eye size={16} />
                      </button>
                    </td>
                  </tr>
                ))}
            </tbody>
          </table>
        </div>

        <div className="billing-pagination d-flex justify-content-between align-items-center mt-3">
          <div className="d-flex align-items-center gap-2">
            <span>{t("Show")}</span>
            <select
              className="form-select-custom"
              value={pageSize}
              onChange={(e) => setPageSize(Number(e.target.value))}
            >
              <option value={10}>10</option>
              <option value={25}>25</option>
            </select>
            <span>
              {t("Showing")} {reviews.length} {t("Results")}
            </span>
          </div>
        </div>
      </div>

      {/* MODAL: REVIEW DETAILS (SCREENSHOT 180643) */}
      {selectedReview && (
        <div className="modal-backdrop-custom">
          <div className="modal-card-custom" style={{ maxWidth: "560px" }}>
            <div className="modal-header-custom d-flex justify-content-between align-items-center">
              <h3>{t("Review Details")}</h3>
              <button
                className="btn-close-custom"
                onClick={() => setSelectedReview(null)}
              >
                <X size={18} />
              </button>
            </div>
            <div className="modal-body-custom">
              <div className="row g-4 mb-4">
                <div className="col-6">
                  <div className="text-secondary text-uppercase small fw-bold mb-2">
                    {t("DOCTOR")}
                  </div>
                  <div className="d-flex align-items-center gap-2">
                    <div className="doctor-avatar-circle">
                      {selectedReview.doctorName.slice(0, 2).toUpperCase()}
                    </div>
                    <div>
                      <div className="fw-semibold">
                        {selectedReview.doctorName}
                      </div>
                      <div className="text-secondary small">
                        {selectedReview.doctorDepartment}
                      </div>
                    </div>
                  </div>
                </div>

                <div className="col-6">
                  <div className="text-secondary text-uppercase small fw-bold mb-2">
                    {t("REVIEWED BY")}
                  </div>
                  <div className="d-flex align-items-center gap-2">
                    <div className="patient-avatar-circle">
                      {selectedReview.patientInitials}
                    </div>
                    <div>
                      <div className="fw-semibold">
                        {selectedReview.patientName}
                      </div>
                      <div className="text-secondary small">
                        {selectedReview.patientEmail}
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              <div className="row g-4 mb-4">
                <div className="col-6">
                  <div className="text-secondary text-uppercase small fw-bold mb-2">
                    {t("RATING")}
                  </div>
                  <div className="d-flex align-items-center text-warning gap-1">
                    {[...Array(selectedReview.rating)].map((_, i) => (
                      <Star key={i} size={18} fill="#f59e0b" color="#f59e0b" />
                    ))}
                  </div>
                </div>

                <div className="col-6">
                  <div className="text-secondary text-uppercase small fw-bold mb-2">
                    {t("STATUS")}
                  </div>
                  <span className="badge-green">{selectedReview.status}</span>
                </div>
              </div>

              <div className="mb-4">
                <div className="text-secondary text-uppercase small fw-bold mb-2">
                  {t("COMMENT")}
                </div>
                <div className="review-comment-card">
                  {selectedReview.comment}
                </div>
              </div>

              <div>
                <div className="text-secondary text-uppercase small fw-bold mb-1">
                  {t("CREATED AT")}
                </div>
                <div className="fw-medium">{selectedReview.createdAt}</div>
              </div>
            </div>

            <div className="modal-footer-custom d-flex justify-content-end">
              <button
                type="button"
                className="btn-action-grey px-4"
                onClick={() => setSelectedReview(null)}
              >
                {t("Cancel")}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
