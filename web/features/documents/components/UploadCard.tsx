import { Card } from '@/shared/ui/Card';

export type UploadCardProps = {
  isDragging: boolean;
  uploading: boolean;
  uploadStatus: string;
  handleDragEnter: (e: React.DragEvent) => void;
  handleDragOver: (e: React.DragEvent) => void;
  handleDragLeave: (e: React.DragEvent) => void;
  handleDrop: (e: React.DragEvent) => void;
  handleFileChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
};

export function UploadCard({
  isDragging,
  uploading,
  uploadStatus,
  handleDragEnter,
  handleDragOver,
  handleDragLeave,
  handleDrop,
  handleFileChange,
}: UploadCardProps) {
  return (
    <Card title="Upload New Document">
      <div
        className={`border-2 border-dashed rounded-lg p-8 text-center transition-colors ${
          isDragging
            ? 'border-slate-900 bg-slate-50'
            : 'border-slate-300 hover:border-slate-400'
        }`}
        onDragEnter={handleDragEnter}
        onDragOver={handleDragOver}
        onDragLeave={handleDragLeave}
        onDrop={handleDrop}
      >
        <input
          type="file"
          accept=".pdf"
          onChange={handleFileChange}
          disabled={uploading}
          className="hidden"
          id="file-upload"
        />
        <div className="space-y-4">
          <div className="text-slate-600">
            <svg
              className="mx-auto h-12 w-12 text-slate-400"
              stroke="currentColor"
              fill="none"
              viewBox="0 0 48 48"
              aria-hidden="true"
            >
              <path
                d="M28 8H12a4 4 0 00-4 4v20m32-12v8m0 0v8a4 4 0 01-4 4H12a4 4 0 01-4-4v-4m32-4l-3.172-3.172a4 4 0 00-5.656 0L28 28M8 32l9.172-9.172a4 4 0 015.656 0L28 28m0 0l4 4m4-24h8m-4-4v8m-12 4h.02"
                strokeWidth={2}
                strokeLinecap="round"
                strokeLinejoin="round"
              />
            </svg>
            <p className="mt-2 text-sm font-medium">
              {isDragging ? 'Drop your PDF file here' : 'Drag and drop your PDF file here'}
            </p>
            <p className="mt-1 text-xs text-slate-500">or</p>
          </div>
          <label
            htmlFor="file-upload"
            className="cursor-pointer inline-block px-6 py-3 bg-slate-900 text-white rounded-lg hover:bg-slate-800 disabled:opacity-50 text-sm font-medium"
          >
            {uploading ? 'Uploading...' : 'Choose PDF File'}
          </label>
        </div>
        {uploadStatus && (
          <p className="mt-4 text-sm text-slate-600">{uploadStatus}</p>
        )}
      </div>
    </Card>
  );
}
