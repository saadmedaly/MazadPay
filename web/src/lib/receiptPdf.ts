import html2canvas from 'html2canvas'
import jsPDF from 'jspdf'

// Convertit un élément HTML déjà monté (le ReceiptTemplate, rendu hors écran) en PDF
// téléchargeable. html2canvas capture le rendu réel du navigateur (texte arabe RTL
// correctement formé), puis jsPDF encapsule l'image capturée dans un document PDF —
// évite tout problème de police/RTL qu'un dessin direct via l'API texte de jsPDF
// aurait posé avec de l'arabe.
export async function downloadReceiptPdf(element: HTMLElement, filename: string): Promise<void> {
  const canvas = await html2canvas(element, {
    scale: 2, // meilleure résolution pour l'impression
    backgroundColor: '#ffffff',
    useCORS: true,
  })

  const imgData = canvas.toDataURL('image/png')
  const pdf = new jsPDF({
    orientation: 'portrait',
    unit: 'px',
    format: [canvas.width, canvas.height],
  })

  pdf.addImage(imgData, 'PNG', 0, 0, canvas.width, canvas.height)
  pdf.save(filename)
}
