'use client';

import { useMemo, useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { AppNav } from '@/shared/ui/AppNav';
import { cardClass } from '@/lib/ui';
import { useAccountOptions } from '@/shared/account-scope/useAccountOptions';
import { useDocuments } from '@/features/documents/hooks/useDocuments';
import { apiClient } from '@/lib/api-client';
import {
  createInstitution,
  updateInstitution,
  deleteInstitution,
} from '@/features/institutions/services/institutionsApi';
import {
  createAccount,
  updateAccount,
  deleteAccount,
} from '@/features/accounts/services/accountsApi';
import {
  createUploadUrl,
  deleteDocument,
  updateDocument,
  enqueueParsing,
} from '@/features/documents/services/documentsApi';
import { ConfirmModal } from '@/features/accounts/components/ConfirmModal';
import { DocumentStatusBadge } from '@/shared/ui/DocumentStatusBadge';
import { formatStatementDateRange } from '@/shared/formatters/date';
import type { Account, Institution, Document } from '@/shared/types/api';
import { Pencil, Plus, Trash2, FileText, Upload, RotateCw, AlertCircle } from 'lucide-react';

const ACCOUNT_TYPES = ['CURRENT', 'SAVINGS', 'CREDIT_CARD'] as const;

function formatAccountType(type?: string): string {
  if (!type) return '—';
  const map: Record<string, string> = {
    CURRENT: 'Current',
    SAVINGS: 'Savings',
    CREDIT_CARD: 'Credit Card',
  };
  return map[type] ?? type;
}

export default function AccountsPage() {
  const queryClient = useQueryClient();
  const { institutions = [], accounts = [], isLoading } = useAccountOptions();

  const [addInstOpen, setAddInstOpen] = useState(false);
  const [editInstOpen, setEditInstOpen] = useState<Institution | null>(null);
  const [addAccOpen, setAddAccOpen] = useState<Institution | null>(null);
  const [editAccOpen, setEditAccOpen] = useState<Account | null>(null);
  const [uploadAccOpen, setUploadAccOpen] = useState<Account | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<
    | { type: 'institution'; item: Institution }
    | { type: 'account'; item: Account }
    | { type: 'document'; item: Document; accountId: string }
    | null
  >(null);
  const [reassignDoc, setReassignDoc] = useState<{ doc: Document; accountId: string } | null>(null);

  const accountsByInstitution = useMemo(() => {
    const byInst = new Map<string, Account[]>();
    const unassigned: Account[] = [];
    for (const acc of accounts) {
      const instId = acc.institution_id ?? '__unassigned__';
      if (instId === '__unassigned__') {
        unassigned.push(acc);
      } else {
        if (!byInst.has(instId)) byInst.set(instId, []);
        byInst.get(instId)!.push(acc);
      }
    }
    return { byInst, unassigned };
  }, [accounts]);

  const institutionOrder = useMemo(() => {
    return [...institutions].sort((a, b) => (a.name ?? '').localeCompare(b.name ?? ''));
  }, [institutions]);

  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: ['institutions'] });
    queryClient.invalidateQueries({ queryKey: ['accounts'] });
    queryClient.invalidateQueries({ queryKey: ['documents'] });
  };

  const createInstMutation = useMutation({
    mutationFn: (name: string) => createInstitution(name),
    onSuccess: () => {
      invalidate();
      setAddInstOpen(false);
    },
  });

  const updateInstMutation = useMutation({
    mutationFn: ({ id, name }: { id: string; name: string }) => updateInstitution(id, name),
    onSuccess: () => {
      invalidate();
      setEditInstOpen(null);
    },
  });

  const deleteInstMutation = useMutation({
    mutationFn: (id: string) => deleteInstitution(id),
    onSuccess: () => {
      invalidate();
      setDeleteTarget(null);
    },
  });

  const createAccMutation = useMutation({
    mutationFn: (payload: Parameters<typeof createAccount>[0]) => createAccount(payload),
    onSuccess: () => {
      invalidate();
      setAddAccOpen(null);
    },
  });

  const updateAccMutation = useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: Parameters<typeof updateAccount>[1] }) =>
      updateAccount(id, payload),
    onSuccess: () => {
      invalidate();
      setEditAccOpen(null);
    },
  });

  const deleteAccMutation = useMutation({
    mutationFn: (id: string) => deleteAccount(id),
    onSuccess: () => {
      invalidate();
      setDeleteTarget(null);
    },
  });

  const deleteDocMutation = useMutation({
    mutationFn: (id: string) => deleteDocument(id),
    onSuccess: () => {
      invalidate();
      setDeleteTarget(null);
    },
  });

  const updateDocMutation = useMutation({
    mutationFn: ({ docId, accountId }: { docId: string; accountId: string | null }) =>
      apiClient.updateDocument(docId, { account_id: accountId ?? undefined }),
    onSuccess: () => {
      invalidate();
      setReassignDoc(null);
    },
  });

  const retryDocMutation = useMutation({
    mutationFn: ({ documentId, gcsUri }: { documentId: string; gcsUri: string }) =>
      enqueueParsing(documentId, gcsUri),
    onSuccess: () => {
      invalidate();
    },
  });

  return (
    <div className="min-h-screen bg-slate-50">
      <AppNav active="accounts" />

      <main className="container mx-auto px-6 py-8">
        <div className="space-y-6">
          <div className="flex items-center justify-between">
            <div className="space-y-1">
              <h1 className="text-3xl font-semibold tracking-tight text-slate-900">Accounts</h1>
              <p className="text-sm text-slate-600">
                Manage institutions, accounts, and their documents
              </p>
            </div>
            <button
              onClick={() => setAddInstOpen(true)}
              className="inline-flex items-center gap-2 px-4 py-2 bg-slate-900 text-white rounded-lg hover:bg-slate-800 text-sm font-medium"
            >
              <Plus size={16} />
              Add Institution
            </button>
          </div>

          {isLoading ? (
            <p className="text-sm text-slate-600">Loading...</p>
          ) : (
            <div className="space-y-6">
              {institutionOrder.map((inst) => {
                const instAccounts = accountsByInstitution.byInst.get(inst.institution_id) ?? [];
                return (
                  <InstitutionCard
                    key={inst.institution_id}
                    institution={inst}
                    accounts={instAccounts}
                    allAccounts={accounts}
                    allInstitutions={institutions}
                    onEditInst={() => setEditInstOpen(inst)}
                    onDeleteInst={() => setDeleteTarget({ type: 'institution', item: inst })}
                    onAddAccount={() => setAddAccOpen(inst)}
                    onEditAccount={(acc) => setEditAccOpen(acc)}
                    onDeleteAccount={(acc) => setDeleteTarget({ type: 'account', item: acc })}
                    onUploadDoc={(acc) => setUploadAccOpen(acc)}
                    onDeleteDoc={(doc, accId) =>
                      setDeleteTarget({ type: 'document', item: doc, accountId: accId })
                    }
                    onReassignDoc={(doc, accId) => setReassignDoc({ doc, accountId: accId })}
                    onRetryDoc={(doc) =>
                      retryDocMutation.mutate({ documentId: doc.document_id, gcsUri: doc.gcs_uri })
                    }
                    retryState={{
                      isPending: retryDocMutation.isPending,
                      documentId: retryDocMutation.variables?.documentId ?? null,
                    }}
                    invalidate={invalidate}
                  />
                );
              })}

              {accountsByInstitution.unassigned.length > 0 && (
                <div className={cardClass}>
                  <h2 className="text-base font-semibold text-slate-900 mb-4">Unassigned Accounts</h2>
                  <AccountList
                    accounts={accountsByInstitution.unassigned}
                    allAccounts={accounts}
                    allInstitutions={institutions}
                    onEdit={(acc) => setEditAccOpen(acc)}
                    onDelete={(acc) => setDeleteTarget({ type: 'account', item: acc })}
                    onUploadDoc={(acc) => setUploadAccOpen(acc)}
                    onDeleteDoc={(doc, accId) =>
                      setDeleteTarget({ type: 'document', item: doc, accountId: accId })
                    }
                    onReassignDoc={(doc, accId) => setReassignDoc({ doc, accountId: accId })}
                    onRetryDoc={(doc) =>
                      retryDocMutation.mutate({ documentId: doc.document_id, gcsUri: doc.gcs_uri })
                    }
                    retryState={{
                      isPending: retryDocMutation.isPending,
                      documentId: retryDocMutation.variables?.documentId ?? null,
                    }}
                    invalidate={invalidate}
                  />
                </div>
              )}

              {institutions.length === 0 && accounts.length === 0 && (
                <p className="text-sm text-slate-500 py-8 text-center">
                  No institutions or accounts yet. Click &quot;Add Institution&quot; to get started.
                </p>
              )}
            </div>
          )}
        </div>
      </main>

      {/* Add Institution Modal */}
      {addInstOpen && (
        <SimpleModal
          title="Add Institution"
          onClose={() => setAddInstOpen(false)}
          onSubmit={(name) => createInstMutation.mutate(name)}
          isPending={createInstMutation.isPending}
          error={createInstMutation.error?.message}
          fields={[{ name: 'name', label: 'Name', type: 'text', required: true }]}
        />
      )}

      {/* Edit Institution Modal */}
      {editInstOpen && (
        <SimpleModal
          title="Edit Institution"
          initialValues={{ name: editInstOpen.name }}
          onClose={() => setEditInstOpen(null)}
          onSubmit={(name) =>
            updateInstMutation.mutate({ id: editInstOpen.institution_id, name })
          }
          isPending={updateInstMutation.isPending}
          error={updateInstMutation.error?.message}
          fields={[{ name: 'name', label: 'Name', type: 'text', required: true }]}
        />
      )}

      {/* Add Account Modal */}
      {addAccOpen && (
        <AccountFormModal
          title="Add Account"
          institutions={institutions}
          initialInstitutionId={addAccOpen.institution_id}
          fixedInstitutionId={addAccOpen.institution_id}
          onClose={() => setAddAccOpen(null)}
          onSubmit={(payload) =>
            createAccMutation.mutate({
              institution_id: addAccOpen.institution_id,
              account_name: payload.account_name || undefined,
              account_number: payload.account_number || undefined,
              currency: payload.currency || undefined,
              account_type: payload.account_type || undefined,
            })
          }
          isPending={createAccMutation.isPending}
          error={createAccMutation.error?.message}
        />
      )}

      {/* Edit Account Modal */}
      {editAccOpen && (
        <AccountFormModal
          title="Edit Account"
          institutions={institutions}
          initialValues={{
            institution_id: editAccOpen.institution_id ?? '',
            account_name: editAccOpen.account_name ?? '',
            account_number: editAccOpen.account_number ?? '',
            currency: editAccOpen.currency ?? 'GBP',
            account_type: editAccOpen.account_type ?? '',
          }}
          onClose={() => setEditAccOpen(null)}
          onSubmit={(payload) =>
            updateAccMutation.mutate({ id: editAccOpen.account_id, payload })
          }
          isPending={updateAccMutation.isPending}
          error={updateAccMutation.error?.message}
        />
      )}

      {/* Upload Document Modal */}
      {uploadAccOpen && (
        <UploadDocumentModal
          accountId={uploadAccOpen.account_id}
          accountName={uploadAccOpen.account_name ?? uploadAccOpen.account_number ?? 'Account'}
          onClose={() => setUploadAccOpen(null)}
          onSuccess={() => {
            invalidate();
            setUploadAccOpen(null);
          }}
        />
      )}

      {/* Reassign Document Modal */}
      {reassignDoc && (
        <ReassignDocumentModal
          document={reassignDoc.doc}
          currentAccountId={reassignDoc.accountId}
          accounts={accounts}
          onClose={() => setReassignDoc(null)}
          onReassign={(accountId) =>
            updateDocMutation.mutate({ docId: reassignDoc.doc.document_id, accountId })
          }
          isPending={updateDocMutation.isPending}
        />
      )}

      {/* Delete Confirm Modals */}
      <ConfirmModal
        show={deleteTarget?.type === 'institution'}
        title="Delete Institution"
        message="This will delete the institution, all its accounts, documents, and transactions. This cannot be undone."
        isPending={deleteInstMutation.isPending}
        error={deleteInstMutation.error?.message ?? null}
        onCancel={() => setDeleteTarget(null)}
        onConfirm={() =>
          deleteTarget?.type === 'institution' && deleteInstMutation.mutate(deleteTarget.item.institution_id)
        }
      />
      <ConfirmModal
        show={deleteTarget?.type === 'account'}
        title="Delete Account"
        message="This will delete the account, all its documents, and transactions. This cannot be undone."
        isPending={deleteAccMutation.isPending}
        error={deleteAccMutation.error?.message ?? null}
        onCancel={() => setDeleteTarget(null)}
        onConfirm={() =>
          deleteTarget?.type === 'account' && deleteAccMutation.mutate(deleteTarget.item.account_id)
        }
      />
      <ConfirmModal
        show={deleteTarget?.type === 'document'}
        title="Delete Document"
        message="This will delete the document and all associated transactions. This cannot be undone."
        isPending={deleteDocMutation.isPending}
        error={deleteDocMutation.error?.message ?? null}
        onCancel={() => setDeleteTarget(null)}
        onConfirm={() =>
          deleteTarget?.type === 'document' && deleteDocMutation.mutate(deleteTarget.item.document_id)
        }
      />
    </div>
  );
}

function InstitutionCard({
  institution,
  accounts,
  allAccounts,
  allInstitutions,
  onEditInst,
  onDeleteInst,
  onAddAccount,
  onEditAccount,
  onDeleteAccount,
  onUploadDoc,
  onDeleteDoc,
  onReassignDoc,
  onRetryDoc,
  retryState,
  invalidate,
}: {
  institution: Institution;
  accounts: Account[];
  allAccounts: Account[];
  allInstitutions: Institution[];
  onEditInst: () => void;
  onDeleteInst: () => void;
  onAddAccount: () => void;
  onEditAccount: (acc: Account) => void;
  onDeleteAccount: (acc: Account) => void;
  onUploadDoc: (acc: Account) => void;
  onDeleteDoc: (doc: Document, accountId: string) => void;
  onReassignDoc: (doc: Document, accountId: string) => void;
  onRetryDoc: (doc: Document) => void;
  retryState: { isPending: boolean; documentId: string | null };
  invalidate: () => void;
}) {
  return (
    <div className={cardClass}>
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-base font-semibold text-slate-900">{institution.name}</h2>
        <div className="flex items-center gap-2">
          <button
            onClick={onAddAccount}
            className="p-2 text-slate-600 hover:bg-slate-100 rounded-lg"
            title="Add account"
          >
            <Plus size={16} />
          </button>
          <button
            onClick={onEditInst}
            className="p-2 text-slate-600 hover:bg-slate-100 rounded-lg"
            title="Edit institution"
          >
            <Pencil size={16} />
          </button>
          <button
            onClick={onDeleteInst}
            className="p-2 text-red-600 hover:bg-red-50 rounded-lg"
            title="Delete institution"
          >
            <Trash2 size={16} />
          </button>
        </div>
      </div>
      <AccountList
        accounts={accounts}
        allAccounts={allAccounts}
        allInstitutions={allInstitutions}
        onEdit={onEditAccount}
        onDelete={onDeleteAccount}
        onUploadDoc={onUploadDoc}
        onDeleteDoc={onDeleteDoc}
        onReassignDoc={onReassignDoc}
        onRetryDoc={onRetryDoc}
        retryState={retryState}
        invalidate={invalidate}
      />
    </div>
  );
}

function AccountList({
  accounts,
  allAccounts,
  allInstitutions,
  onEdit,
  onDelete,
  onUploadDoc,
  onDeleteDoc,
  onReassignDoc,
  onRetryDoc,
  retryState,
  invalidate,
}: {
  accounts: Account[];
  allAccounts: Account[];
  allInstitutions: Institution[];
  onEdit: (acc: Account) => void;
  onDelete: (acc: Account) => void;
  onUploadDoc: (acc: Account) => void;
  onDeleteDoc: (doc: Document, accountId: string) => void;
  onReassignDoc: (doc: Document, accountId: string) => void;
  onRetryDoc: (doc: Document) => void;
  retryState: { isPending: boolean; documentId: string | null };
  invalidate: () => void;
}) {
  return (
    <div className="space-y-4">
      {accounts.length === 0 ? (
        <p className="text-sm text-slate-500 py-2">No accounts</p>
      ) : (
        accounts.map((acc) => (
          <AccountRow
            key={acc.account_id}
            account={acc}
            allAccounts={allAccounts}
            allInstitutions={allInstitutions}
            onEdit={() => onEdit(acc)}
            onDelete={() => onDelete(acc)}
            onUploadDoc={() => onUploadDoc(acc)}
            onDeleteDoc={(doc) => onDeleteDoc(doc, acc.account_id)}
            onReassignDoc={(doc) => onReassignDoc(doc, acc.account_id)}
            onRetryDoc={onRetryDoc}
            retryState={retryState}
            invalidate={invalidate}
          />
        ))
      )}
    </div>
  );
}

function AccountRow({
  account,
  allAccounts,
  allInstitutions,
  onEdit,
  onDelete,
  onUploadDoc,
  onDeleteDoc,
  onReassignDoc,
  onRetryDoc,
  retryState,
  invalidate,
}: {
  account: Account;
  allAccounts: Account[];
  allInstitutions: Institution[];
  onEdit: () => void;
  onDelete: () => void;
  onUploadDoc: () => void;
  onDeleteDoc: (doc: Document) => void;
  onReassignDoc: (doc: Document) => void;
  onRetryDoc: (doc: Document) => void;
  retryState: { isPending: boolean; documentId: string | null };
  invalidate: () => void;
}) {
  const { data: documents = [], isLoading } = useDocuments({
    scope: { mode: 'account', accountId: account.account_id },
  });

  return (
    <div className="border border-slate-200 rounded-lg p-4 bg-slate-50/50">
      <div className="flex items-center justify-between mb-2">
        <div className="flex items-center gap-4">
          <span className="text-sm font-medium text-slate-900">
            {account.account_name || account.account_number || 'Unnamed'}
          </span>
          <span className="text-xs text-slate-500">{formatAccountType(account.account_type)}</span>
          <span className="text-xs text-slate-500 tabular-nums">{account.account_number || '—'}</span>
          <span className="text-xs text-slate-500">{account.currency || '—'}</span>
        </div>
        <div className="flex items-center gap-1">
          <button
            onClick={onUploadDoc}
            className="p-1.5 text-slate-600 hover:bg-slate-200 rounded"
            title="Upload document"
          >
            <Upload size={14} />
          </button>
          <button onClick={onEdit} className="p-1.5 text-slate-600 hover:bg-slate-200 rounded" title="Edit">
            <Pencil size={14} />
          </button>
          <button onClick={onDelete} className="p-1.5 text-red-600 hover:bg-red-100 rounded" title="Delete">
            <Trash2 size={14} />
          </button>
        </div>
      </div>
      <div className="mt-3">
        <p className="text-[11px] font-medium text-slate-500 uppercase tracking-wider mb-2">
          Documents
        </p>
        {isLoading ? (
          <p className="text-xs text-slate-500">Loading...</p>
        ) : documents.length === 0 ? (
          <div className="flex flex-col items-start gap-2">
            <p className="text-xs text-slate-500">No documents yet</p>
            <button
              onClick={onUploadDoc}
              className="text-xs font-medium text-slate-700 hover:text-slate-900 hover:underline flex items-center gap-1"
            >
              <Upload size={12} />
              Upload statement
            </button>
          </div>
        ) : (
          <ul className="space-y-1">
            {documents.map((doc: Document) => (
              <li
                key={doc.document_id}
                className="flex items-center justify-between text-sm py-1.5 px-2 rounded bg-white border border-slate-100"
              >
                <div className="flex items-center gap-2 min-w-0">
                  <FileText size={14} className="text-slate-400 shrink-0" />
                  <div className="flex flex-col min-w-0">
                    {doc.statement_start_date && doc.statement_end_date ? (
                      <>
                        <span className="font-medium text-slate-900">
                          {formatStatementDateRange(doc.statement_start_date, doc.statement_end_date)}
                        </span>
                        {(doc.original_filename || doc.document_id) && (
                          <span className="text-xs text-slate-500 truncate">
                            {doc.original_filename || doc.document_id}
                          </span>
                        )}
                      </>
                    ) : (
                      <span className="font-medium text-slate-900">
                        {doc.original_filename || doc.document_id}
                      </span>
                    )}
                  </div>
                  <DocumentStatusBadge status={doc.parsing_status} />
                  {doc.parsing_status === 'FAILED' && doc.error_message && (
                      <span
                        title={doc.error_message}
                        className="text-red-500 cursor-help"
                      >
                        <AlertCircle size={14} />
                      </span>
                    )}
                </div>
                <div className="flex items-center gap-1">
                  {doc.parsing_status === 'FAILED' && (
                    <button
                      onClick={() => onRetryDoc(doc)}
                      disabled={
                        retryState.isPending &&
                        retryState.documentId === doc.document_id
                      }
                      className="text-xs text-slate-600 hover:underline flex items-center gap-1 disabled:opacity-50"
                      title="Retry parsing"
                    >
                      <RotateCw
                        size={12}
                        className={
                          retryState.isPending &&
                          retryState.documentId === doc.document_id
                            ? 'animate-spin'
                            : ''
                        }
                      />
                      Retry
                    </button>
                  )}
                  <button
                    onClick={() => onReassignDoc(doc)}
                    className="text-xs text-slate-600 hover:underline"
                  >
                    Reassign
                  </button>
                  <button
                    onClick={() => onDeleteDoc(doc)}
                    className="p-1 text-red-600 hover:bg-red-50 rounded"
                  >
                    <Trash2 size={12} />
                  </button>
                </div>
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  );
}

function SimpleModal({
  title,
  initialValues = {},
  onClose,
  onSubmit,
  isPending,
  error,
  fields,
}: {
  title: string;
  initialValues?: Record<string, string>;
  onClose: () => void;
  onSubmit: (value: string) => void;
  isPending: boolean;
  error?: string;
  fields: { name: string; label: string; type: string; required?: boolean }[];
}) {
  const [values, setValues] = useState<Record<string, string>>(initialValues);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const nameField = fields.find((f) => f.name === 'name');
    if (nameField) onSubmit(values[nameField.name] ?? '');
  };

  return (
    <div className="fixed inset-0 bg-black/40 flex items-center justify-center z-50">
      <div className="bg-white rounded-2xl shadow-lg max-w-md w-full mx-4">
        <form onSubmit={handleSubmit}>
          <div className="p-6">
            <h3 className="text-lg font-semibold text-slate-900 mb-4">{title}</h3>
            {fields.map((f) => (
              <div key={f.name} className="mb-4">
                <label className="block text-sm font-medium text-slate-700 mb-1">{f.label}</label>
                <input
                  type={f.type}
                  value={values[f.name] ?? ''}
                  onChange={(e) => setValues((v) => ({ ...v, [f.name]: e.target.value }))}
                  required={f.required}
                  className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-slate-900 focus:border-transparent"
                />
              </div>
            ))}
            {error && <p className="text-sm text-red-600 mb-2">{error}</p>}
          </div>
          <div className="bg-slate-50 px-6 py-4 flex gap-3 justify-end rounded-b-2xl">
            <button type="button" onClick={onClose} className="px-4 py-2 text-sm font-medium text-slate-700">
              Cancel
            </button>
            <button
              type="submit"
              disabled={isPending}
              className="px-4 py-2 text-sm font-medium bg-slate-900 text-white rounded-lg disabled:opacity-50"
            >
              {isPending ? 'Saving...' : 'Save'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

function AccountFormModal({
  title,
  institutions,
  initialInstitutionId,
  initialValues,
  fixedInstitutionId,
  onClose,
  onSubmit,
  isPending,
  error,
}: {
  title: string;
  institutions: Institution[];
  initialInstitutionId?: string;
  initialValues?: Record<string, string>;
  fixedInstitutionId?: string;
  onClose: () => void;
  onSubmit: (payload: Record<string, string>) => void;
  isPending: boolean;
  error?: string;
}) {
  const [values, setValues] = useState<Record<string, string>>({
    institution_id: initialInstitutionId ?? fixedInstitutionId ?? '',
    account_name: '',
    account_number: '',
    currency: 'GBP',
    account_type: '',
    ...initialValues,
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSubmit(values);
  };

  const showInstitutionField = !fixedInstitutionId;

  return (
    <div className="fixed inset-0 bg-black/40 flex items-center justify-center z-50">
      <div className="bg-white rounded-2xl shadow-lg max-w-md w-full mx-4 max-h-[90vh] overflow-y-auto">
        <form onSubmit={handleSubmit}>
          <div className="p-6">
            <h3 className="text-lg font-semibold text-slate-900 mb-4">{title}</h3>
            <div className="space-y-4">
              {showInstitutionField && (
              <div>
                <label className="block text-sm font-medium text-slate-700 mb-1">Institution</label>
                <select
                  value={values.institution_id}
                  onChange={(e) => setValues((v) => ({ ...v, institution_id: e.target.value }))}
                  required
                  className="w-full px-3 py-2 border border-slate-300 rounded-lg"
                >
                  <option value="">Select...</option>
                  {institutions.map((i) => (
                    <option key={i.institution_id} value={i.institution_id}>
                      {i.name}
                    </option>
                  ))}
                </select>
              </div>
              )}
              <div>
                <label className="block text-sm font-medium text-slate-700 mb-1">Account name</label>
                <input
                  type="text"
                  value={values.account_name}
                  onChange={(e) => setValues((v) => ({ ...v, account_name: e.target.value }))}
                  className="w-full px-3 py-2 border border-slate-300 rounded-lg"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-slate-700 mb-1">Account number</label>
                <input
                  type="text"
                  value={values.account_number}
                  onChange={(e) => setValues((v) => ({ ...v, account_number: e.target.value }))}
                  className="w-full px-3 py-2 border border-slate-300 rounded-lg"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-slate-700 mb-1">Currency</label>
                <input
                  type="text"
                  value={values.currency}
                  onChange={(e) => setValues((v) => ({ ...v, currency: e.target.value }))}
                  placeholder="GBP"
                  className="w-full px-3 py-2 border border-slate-300 rounded-lg"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-slate-700 mb-1">Account type</label>
                <select
                  value={values.account_type}
                  onChange={(e) => setValues((v) => ({ ...v, account_type: e.target.value }))}
                  className="w-full px-3 py-2 border border-slate-300 rounded-lg"
                >
                  <option value="">—</option>
                  {ACCOUNT_TYPES.map((t) => (
                    <option key={t} value={t}>
                      {formatAccountType(t)}
                    </option>
                  ))}
                </select>
              </div>
            </div>
            {error && <p className="text-sm text-red-600 mt-2">{error}</p>}
          </div>
          <div className="bg-slate-50 px-6 py-4 flex gap-3 justify-end rounded-b-2xl">
            <button type="button" onClick={onClose} className="px-4 py-2 text-sm font-medium text-slate-700">
              Cancel
            </button>
            <button
              type="submit"
              disabled={isPending}
              className="px-4 py-2 text-sm font-medium bg-slate-900 text-white rounded-lg disabled:opacity-50"
            >
              {isPending ? 'Saving...' : 'Save'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

function UploadDocumentModal({
  accountId,
  accountName,
  onClose,
  onSuccess,
}: {
  accountId: string;
  accountName: string;
  onClose: () => void;
  onSuccess: () => void;
}) {
  const [uploading, setUploading] = useState(false);
  const [status, setStatus] = useState('');
  const [file, setFile] = useState<File | null>(null);

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const f = e.target.files?.[0];
    if (f) setFile(f);
  };

  const handleUpload = async () => {
    if (!file) return;
    setUploading(true);
    setStatus('Creating upload URL...');
    try {
      const { upload_url, document_id, gcs_uri } = await createUploadUrl(file.name, accountId);
      setStatus('Uploading...');
      const apiBase = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
      const url = upload_url.startsWith('http') ? upload_url : `${apiBase}${upload_url}`;
      const res = await fetch(url, { method: 'POST', body: file, headers: { 'Content-Type': file.type } });
      if (!res.ok) throw new Error('Upload failed');
      setStatus('Enqueueing parsing...');
      await enqueueParsing(document_id, gcs_uri);
      setStatus('Done!');
      onSuccess();
    } catch (err) {
      setStatus(`Error: ${err instanceof Error ? err.message : 'Unknown'}`);
    } finally {
      setUploading(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black/40 flex items-center justify-center z-50">
      <div className="bg-white rounded-2xl shadow-lg max-w-md w-full mx-4">
        <div className="p-6">
          <h3 className="text-lg font-semibold text-slate-900 mb-2">Upload Document</h3>
          <p className="text-sm text-slate-600 mb-4">Add a statement to {accountName}</p>
          <input
            type="file"
            accept=".pdf"
            onChange={handleFileChange}
            disabled={uploading}
            className="block w-full text-sm text-slate-500 file:mr-4 file:py-2 file:px-4 file:rounded-lg file:border-0 file:bg-slate-100 file:text-slate-700"
          />
          {status && <p className="mt-2 text-sm text-slate-600">{status}</p>}
        </div>
        <div className="bg-slate-50 px-6 py-4 flex gap-3 justify-end rounded-b-2xl">
          <button onClick={onClose} disabled={uploading} className="px-4 py-2 text-sm font-medium text-slate-700">
            Cancel
          </button>
          <button
            onClick={handleUpload}
            disabled={!file || uploading}
            className="px-4 py-2 text-sm font-medium bg-slate-900 text-white rounded-lg disabled:opacity-50"
          >
            {uploading ? 'Uploading...' : 'Upload'}
          </button>
        </div>
      </div>
    </div>
  );
}

function ReassignDocumentModal({
  document,
  currentAccountId,
  accounts,
  onClose,
  onReassign,
  isPending,
}: {
  document: Document;
  currentAccountId: string;
  accounts: Account[];
  onClose: () => void;
  onReassign: (accountId: string | null) => void;
  isPending: boolean;
}) {
  const [selectedId, setSelectedId] = useState<string>(currentAccountId);

  return (
    <div className="fixed inset-0 bg-black/40 flex items-center justify-center z-50">
      <div className="bg-white rounded-2xl shadow-lg max-w-md w-full mx-4">
        <div className="p-6">
          <h3 className="text-lg font-semibold text-slate-900 mb-2">Reassign Document</h3>
          <p className="text-sm text-slate-600 mb-4">
            {document.original_filename || document.document_id}
          </p>
          <label className="block text-sm font-medium text-slate-700 mb-1">Assign to account</label>
          <select
            value={selectedId}
            onChange={(e) => setSelectedId(e.target.value)}
            className="w-full px-3 py-2 border border-slate-300 rounded-lg"
          >
            <option value="">Unassigned</option>
            {accounts.map((a) => (
              <option key={a.account_id} value={a.account_id}>
                {a.account_name || a.account_number || a.account_id} ({a.currency || '—'})
              </option>
            ))}
          </select>
        </div>
        <div className="bg-slate-50 px-6 py-4 flex gap-3 justify-end rounded-b-2xl">
          <button onClick={onClose} disabled={isPending} className="px-4 py-2 text-sm font-medium text-slate-700">
            Cancel
          </button>
          <button
            onClick={() => onReassign(selectedId ? selectedId : null)}
            disabled={isPending}
            className="px-4 py-2 text-sm font-medium bg-slate-900 text-white rounded-lg disabled:opacity-50"
          >
            {isPending ? 'Saving...' : 'Reassign'}
          </button>
        </div>
      </div>
    </div>
  );
}
