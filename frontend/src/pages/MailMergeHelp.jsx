import { ArrowLeftIcon } from '@heroicons/react/24/outline'
import { useNavigate } from 'react-router-dom'

export default function MailMergeHelp() {
  const navigate = useNavigate()

  return (
    <div className="min-h-screen bg-gray-50 py-8">
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="bg-white shadow rounded-lg">
          <div className="px-6 py-4 border-b border-gray-200">
            <div className="flex items-center">
              <button
                onClick={() => navigate(-1)}
                className="mr-4 p-2 text-gray-400 hover:text-gray-600"
              >
                <ArrowLeftIcon className="h-5 w-5" />
              </button>
              <h1 className="text-2xl font-bold text-gray-900">Gmail Mail Merge Guide</h1>
            </div>
          </div>
          
          <div className="px-6 py-6">
            <div className="prose max-w-none">
              <p className="text-lg text-gray-600 mb-6">
                Learn how to use Gmail mail merge with Google Sheets to send personalized student invitations using your class encryption key.
              </p>

              <h2 className="text-xl font-semibold text-gray-900 mb-4">What You'll Need</h2>
              <ul className="list-disc list-inside mb-6 space-y-2">
                <li>Your class encryption key (displayed on your class screen)</li>
                <li>Gmail account</li>
                <li>Google Sheets access</li>
                <li>Student information (names, parent emails)</li>
              </ul>

              <h2 className="text-xl font-semibold text-gray-900 mb-4">Step-by-Step Instructions</h2>
              
              <div className="space-y-6">
                <div className="border-l-4 border-blue-500 pl-4">
                  <h3 className="font-semibold text-gray-900">Step 1: Access the Template</h3>
                  <p className="text-gray-600 mt-2">
                    Open our pre-configured Google Sheets template that includes the mail merge formulas:
                  </p>
                  <a
                    href="https://docs.google.com/spreadsheets/d/1lnSAl7tZdBeC8vN5E0NyiN_tmYDDUc6ZQHQDDM1LNRY/edit?gid=0#gid=0"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="inline-flex items-center mt-2 px-4 py-2 bg-green-600 text-white rounded hover:bg-green-700"
                  >
                    Open Google Sheets Template
                  </a>
                </div>

                <div className="border-l-4 border-blue-500 pl-4">
                  <h3 className="font-semibold text-gray-900">Step 2: Make a Copy</h3>
                  <p className="text-gray-600 mt-2">
                    Click "File" → "Make a copy" to create your own version of the template.
                  </p>
                </div>

                <div className="border-l-4 border-blue-500 pl-4">
                  <h3 className="font-semibold text-gray-900">Step 3: Add Your Encryption Key</h3>
                  <p className="text-gray-600 mt-2">
                    In the spreadsheet, find the "ENCRYPTION_KEY" cell and paste your class encryption key from the teacher dashboard.
                  </p>
                </div>

                <div className="border-l-4 border-blue-500 pl-4">
                  <h3 className="font-semibold text-gray-900">Step 4: Add Student Information</h3>
                  <p className="text-gray-600 mt-2">
                    Fill in the student information in the provided columns:
                  </p>
                  <ul className="list-disc list-inside mt-2 ml-4 space-y-1">
                    <li>Student Name</li>
                    <li>Parent Email</li>
                    <li>Any additional information as needed</li>
                  </ul>
                </div>

                <div className="border-l-4 border-blue-500 pl-4">
                  <h3 className="font-semibold text-gray-900">Step 5: Install Mail Merge Add-on</h3>
                  <p className="text-gray-600 mt-2">
                    In Google Sheets, go to "Extensions" → "Add-ons" → "Get add-ons" and search for "Mail Merge" (we recommend "Mail Merge with Attachments").
                  </p>
                </div>

                <div className="border-l-4 border-blue-500 pl-4">
                  <h3 className="font-semibold text-gray-900">Step 6: Create Email Template</h3>
                  <p className="text-gray-600 mt-2">
                    Create a Gmail draft with your invitation email. Use the pre-written template below:
                  </p>
                </div>
              </div>

              <h2 className="text-xl font-semibold text-gray-900 mb-4 mt-8">Email Template</h2>
              <div className="bg-gray-100 p-4 rounded-lg mb-6">
                <div className="font-mono text-sm">
                  <div className="mb-2"><strong>Subject:</strong> Invitation to Join {'{{Student Name}}'}'s Reading Class</div>
                  <div className="whitespace-pre-line">
{`Dear Parent/Guardian,

You're invited to join your child ${'{{Student Name}}'}'s reading tracking class!

Click the link below to get started:
${'{{Invitation Link}}'}

This link is personalized for ${'{{Student Name}}'} and will help you track their reading progress throughout the school year.

Best regards,
${'{{Teacher Name}}'}
${'{{School Name}}'}`}
                  </div>
                </div>
              </div>

              <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4 mb-6">
                <h3 className="font-semibold text-yellow-800 mb-2">Important Notes:</h3>
                <ul className="list-disc list-inside text-yellow-700 space-y-1">
                  <li>The invitation links are automatically generated using your encryption key</li>
                  <li>Each link is unique to the student and cannot be shared</li>
                  <li>Links expire after 30 days for security</li>
                  <li>Keep your encryption key secure and don't share it</li>
                </ul>
              </div>

              <h2 className="text-xl font-semibold text-gray-900 mb-4">Troubleshooting</h2>
              <div className="space-y-4">
                <div>
                  <h4 className="font-semibold text-gray-900">Links not working?</h4>
                  <p className="text-gray-600">Make sure you've correctly pasted your encryption key in the designated cell.</p>
                </div>
                <div>
                  <h4 className="font-semibold text-gray-900">Mail merge not sending?</h4>
                  <p className="text-gray-600">Check that your Gmail account has permission to send emails through the add-on.</p>
                </div>
                <div>
                  <h4 className="font-semibold text-gray-900">Need more help?</h4>
                  <p className="text-gray-600">Contact your system administrator or refer to the Google Sheets mail merge documentation.</p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}