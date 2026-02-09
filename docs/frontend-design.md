# Frontend Design

UI/UX design inspired by Nihon Falcom's Trails series quest menus.

## Design Philosophy

Quest Todo's frontend takes direct inspiration from the quest journal system in the Trails series, featuring:
- **Journal/Book aesthetic** with ornate borders and parchment textures
- **Quest-style organization** with main tasks and side tasks
- **Priority system** using star ratings (★★★★★)
- **Deadline indicators** with color-coded urgency (short/medium/long)
- **Objective tracking** with checkboxes and progress bars
- **Reward system** for gamification

## Visual Reference

Based on the provided screenshot, key design elements include:
- Two-panel layout (list + details)
- Parchment/paper texture background
- Tab navigation at top
- Priority stars display
- Client/Category labels
- Deadline indicators
- Objectives checklist with star markers (★/☆)
- Status indicators

---

## Layout Structure

### Overall Application Window

```
┌──────────────────────────────────────────────────────────┐
│  [LT]  ┌──────┬───────┬───────┬───────┐  [RT]            │
│        │ NOTES│ HISTORY│ REQUESTS │ REQUESTS │            │
│        │History│Rolent │  Bose   │                        │
│        └──────┴───────┴───────┴───────┘                   │
├────────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌────────────────────────────────────┐ │
│  │              │  │                                    │ │
│  │  Task List   │  │        Task Details                │ │
│  │              │  │                                    │ │
│  │  [Left Panel]│  │      [Right Panel]                 │ │
│  │              │  │                                    │ │
│  │              │  │                                    │ │
│  │              │  │                                    │ │
│  └──────────────┘  └────────────────────────────────────┘ │
└──────────────────────────────────────────────────────────┘
```

### Decorative Elements
- Ornate corner decorations (medieval/fantasy style)
- Metal ring/binding elements (simulating a book)
- Subtle parchment texture overlay
- Drop shadows for depth
- Page turn animations (optional)

---

## Left Panel - Task List

### Layout

```
┌─────────────────────────┐
│  Main Quest             │
│  ━━━━━━━━━━━━━━━━━━━    │
│  □ Task Title 1         │
│  □ Task Title 2         │
│                         │
│  Sub Quest              │
│  ━━━━━━━━━━━━━━━━━━━    │
│  ◎ Task Title 3         │
│  ◎ Task Title 4         │
│  ▶ Task Title 5 (Active)│
│  ◎ Task Title 6         │
│                         │
└─────────────────────────┘
```

### Elements

**Category Headers:**
- Font: Serif, bold, 18px
- Color: Teal/blue-green (#3A7F8F)
- Underline or decorative divider
- Text: "Main Quest" / "Sub Quest" (or "Main Tasks" / "Side Tasks")

**Task Rows:**
- Font: Serif, 16px
- Height: 32-40px
- Padding: 8px
- Background: Transparent (hover: light gold #F4E4A6)
- Selected: Gold background (#E8D67A)

**Status Icons:**
- `□` Checkbox for incomplete main tasks
- `☑` Checked box for complete
- `◎` Circle for side quests
- `⦿` Filled circle for complete side quests
- `▶` Arrow indicator for currently selected

**Hover Effect:**
- Subtle background color change
- Slight scale transform (1.02x)
- Cursor: pointer

---

## Right Panel - Task Details

### Layout

```
┌────────────────────────────────────────────┐
│  Escort Escapade           [Priority] ★★★☆ │
│                                            │
│  Client: Hart           Deadline: --       │
│  Reward: 1500 Mira      BP Earned: 4(+1)  │
│                                            │
│  ────────────────────────────────────────  │
│                                            │
│  Looking for someone to protect me on the  │
│  way to Krone Pass Checkpoint. If         │
│  available, please meet me at Frieden      │
│  Hotel.                                    │
│                                            │
│  ★ Sounds like we just need to escort     │
│    them to the Krone Pass Checkpoint.     │
│  ★ Agreed to meet the client at Bose's    │
│    West Exit.                              │
│  ★ Krone Pass is located far west along   │
│    the West Bose Highway.                 │
│  ☆ Additional objective here              │
│                                            │
│  ────────────────────────────────────────  │
│                                            │
│  Status: Reported                          │
└────────────────────────────────────────────┘
```

### Elements

**Title Section:**
- Font: Serif, bold, 24px
- Color: Teal (#3A7F8F)
- Aligned left with priority stars on right

**Priority Stars:**
- Filled star: ★ (gold #E8B84D)
- Empty star: ☆ (gray #C0C0C0)
- Size: 20px
- Rating: 1-5 stars

**Metadata Row:**
```
Client: Hart              Deadline: Short (2 days)
Reward: 50 points         BP: 5(+2)
```
- Font: Sans-serif, 14px
- Color: Dark gray (#5A5A5A)
- Layout: Two columns or flexible grid
- Labels in lighter color, values in darker

**Divider:**
- Horizontal line or decorative border
- Color: Light brown (#C0A080)
- Margin: 16px vertical

**Description:**
- Font: Serif, 16px, line-height 1.6
- Color: Dark text (#3A3A3A)
- Max-width for readability
- Padding: 16px vertical

**Objectives List:**
```
★ Completed objective (filled star)
☆ Incomplete objective (empty star)
```
- Font: Serif, 15px
- Line-height: 1.8
- Filled star (★) for completed
- Empty star (☆) for incomplete
- Indent: 8px after star
- Hover: Highlight entire line
- Click: Toggle completion

**Progress Indicator:**
```
Progress: ███████░░░ 3/4 Complete
```
- Visual bar showing completion
- Text: "3/4 objectives complete"

**Status Footer:**
- Font: Sans-serif, 14px
- Color: Based on status:
  - Active: Blue (#3A7F8F)
  - In Progress: Orange (#E8A958)
  - Complete: Green (#6BA573)
  - Failed: Red (#D94A4A)

---

## Top Navigation Tabs

### Tab Bar

```
┌─────┬─────────┬──────────┬─────────┐
│NOTES│ HISTORY │ REQUESTS │REQUESTS │
│History│ Rolent │   Bose   │         │
└─────┴─────────┴──────────┴─────────┘
  Active    Inactive   Inactive
```

**Tab Styles:**
- **Active Tab:**
  - Background: Light parchment (#F5F1E8)
  - Border: Brown outline
  - Font: Bold
  - Slightly raised appearance

- **Inactive Tab:**
  - Background: Darker parchment (#E0D8C8)
  - Border: Subtle outline
  - Font: Regular
  - Slightly recessed appearance

**Tab Labels:**
- "Active" or "Notes" - Current tasks
- "History" - Completed tasks
- "Projects" or category names - Organizational views

---

## Color Scheme

### Primary Colors

| Element | Color | Hex | Usage |
|---------|-------|-----|-------|
| Background | Warm Parchment | #F5F1E8 | Main background |
| Text Primary | Dark Brown | #3A3A3A | Body text |
| Text Secondary | Teal | #3A7F8F | Headers, titles |
| Accent | Deep Burgundy | #5C3A3A | Borders, decorations |
| Selection | Gold | #E8D67A | Selected items |
| Hover | Light Gold | #F4E4A6 | Hover states |

### Status Colors

| Status | Color | Hex |
|--------|-------|-----|
| Short Deadline | Red | #D94A4A |
| Medium Deadline | Orange | #E8A958 |
| Long Deadline | Green | #6BA573 |
| Complete | Gray-Green | #A8A8A8 |
| Failed | Dark Red | #B83A3A |

### Priority Stars

| Type | Color | Hex |
|------|-------|-----|
| Filled Star | Gold | #E8B84D |
| Empty Star | Silver Gray | #C0C0C0 |

---

## Typography

### Font Families

**Primary (Serif):**
- macOS: "Crimson Text", "Georgia", "Times New Roman"
- Usage: Headers, body text, task descriptions
- Character: Classical, readable, book-like

**Secondary (Sans-serif):**
- macOS: "SF Pro Text", "Helvetica Neue"
- Usage: Metadata, labels, UI elements
- Character: Clean, modern, readable

### Font Sizes

| Element | Size | Weight | Line Height |
|---------|------|--------|-------------|
| Task List Header | 18px | Bold | 1.4 |
| Task Title (List) | 16px | Regular | 1.5 |
| Detail Panel Title | 24px | Bold | 1.3 |
| Description Text | 16px | Regular | 1.6 |
| Objectives | 15px | Regular | 1.8 |
| Metadata Labels | 14px | Regular | 1.5 |
| Tab Labels | 14px | Bold | 1.2 |

---

## Interactive Elements

### Task Row (List)

**States:**
```
Normal:     □ Task Title
Hover:      □ Task Title  (light gold background)
Selected:   ▶ Task Title  (gold background)
Complete:   ☑ Task Title  (gray text, strikethrough)
```

**Interactions:**
- Click: Select and show details
- Right-click: Context menu (edit, delete, etc.)
- Double-click: Quick edit
- Drag: Reorder tasks

### Objectives (Details)

**States:**
```
Incomplete: ☆ Objective text
Complete:   ★ Objective text
Hover:      ☆ Objective text  (underline)
```

**Interactions:**
- Click: Toggle completion
- Smooth animation on toggle (star fill/empty)
- Update progress bar in real-time

### Buttons

**Primary Button:**
- Background: Teal (#3A7F8F)
- Text: White
- Border: None
- Padding: 10px 20px
- Border-radius: 4px
- Hover: Darker teal (#2E6B7A)

**Secondary Button:**
- Background: Transparent
- Text: Teal (#3A7F8F)
- Border: 1px solid teal
- Padding: 10px 20px
- Border-radius: 4px
- Hover: Light teal background

---

## Animations

### Page Transitions
- Fade in/out: 200ms
- Slide animations for tab changes: 300ms ease-in-out

### Task Selection
- Background color transition: 150ms
- Smooth scroll to selected item

### Objective Toggle
- Star fill animation: 200ms
- Scale pulse effect: 1.0 → 1.2 → 1.0 (300ms)

### Task Creation
- Fade in new task: 300ms
- Slide down from top

### Loading States
- Skeleton screens with shimmer effect
- Spinner for long operations

---

## Responsive Behavior

### Desktop (Wide Screen)
- Side-by-side panels (30% list / 70% details)
- All tabs visible

### Desktop (Medium)
- Side-by-side panels (35% list / 65% details)
- Tab overflow with scroll

### Tablet
- Single panel view
- Slide transition between list and details
- Back button to return to list

### Mobile (Future)
- Full-screen panels
- Bottom navigation
- Simplified layout

---

## Components Breakdown

### 1. TaskListView

```swift
struct TaskListView: View {
    @ObservedObject var viewModel: TaskListViewModel
    @Binding var selectedTaskId: String?

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 0) {
                ForEach(viewModel.categories) { category in
                    CategoryHeaderView(category: category)

                    ForEach(viewModel.tasks(for: category)) { task in
                        TaskRowView(
                            task: task,
                            isSelected: selectedTaskId == task.id,
                            onTap: { selectedTaskId = task.id }
                        )
                    }
                }
            }
        }
        .background(ParchmentTexture())
    }
}
```

### 2. TaskDetailView

```swift
struct TaskDetailView: View {
    let task: Task
    @ObservedObject var viewModel: TaskDetailViewModel

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 16) {
                // Title & Priority
                HStack {
                    Text(task.title)
                        .font(.custom("Crimson Text", size: 24))
                        .fontWeight(.bold)

                    Spacer()

                    PriorityStars(priority: task.priority)
                }

                // Metadata
                MetadataRow(task: task)

                Divider()

                // Description
                Text(task.description)
                    .font(.custom("Crimson Text", size: 16))
                    .lineSpacing(8)

                // Objectives
                ObjectivesListView(
                    objectives: task.objectives,
                    onToggle: viewModel.toggleObjective
                )

                Divider()

                // Status
                StatusLabel(status: task.status)
            }
            .padding()
        }
        .background(ParchmentTexture())
    }
}
```

### 3. PriorityStars

```swift
struct PriorityStars: View {
    let priority: Int

    var body: some View {
        HStack(spacing: 2) {
            ForEach(1...5, id: \.self) { index in
                Image(systemName: index <= priority ? "star.fill" : "star")
                    .foregroundColor(index <= priority ? .gold : .silverGray)
                    .font(.system(size: 18))
            }
        }
    }
}
```

### 4. DeadlineLabel

```swift
struct DeadlineLabel: View {
    let deadline: Deadline

    var body: some View {
        HStack(spacing: 4) {
            Text("Deadline:")
                .foregroundColor(.secondary)

            Text(deadline.displayText)
                .fontWeight(.semibold)
                .foregroundColor(deadline.color)
        }
        .font(.system(size: 14))
    }
}

extension Deadline {
    var color: Color {
        switch type {
        case .short: return .red
        case .medium: return .orange
        case .long: return .green
        case .none: return .gray
        }
    }

    var displayText: String {
        if let date = date {
            // Format as "2 days" or specific date
            return formatRelativeDate(date)
        }
        return type.rawValue.capitalized
    }
}
```

---

## Custom Views & Modifiers

### Parchment Background

```swift
struct ParchmentTexture: View {
    var body: some View {
        Rectangle()
            .fill(Color.parchment)
            .overlay(
                Image("parchment-texture")
                    .resizable(resizingMode: .tile)
                    .opacity(0.3)
            )
    }
}
```

### Journal Frame

```swift
struct JournalFrame: ViewModifier {
    func body(content: Content) -> some View {
        content
            .padding(20)
            .background(
                RoundedRectangle(cornerRadius: 8)
                    .fill(Color.parchment)
                    .shadow(radius: 10)
            )
            .overlay(
                RoundedRectangle(cornerRadius: 8)
                    .stroke(Color.burgundy, lineWidth: 2)
            )
            .overlay(cornerDecorations)
    }

    var cornerDecorations: some View {
        // Ornate corner decorations
        // Implementation using Path or Images
    }
}
```

---

## Assets & Resources

### Required Images
- `parchment-texture.png` - Subtle paper texture
- `corner-ornament-tl.png` - Top-left corner decoration
- `corner-ornament-tr.png` - Top-right corner decoration
- `corner-ornament-bl.png` - Bottom-left corner decoration
- `corner-ornament-br.png` - Bottom-right corner decoration
- `divider-ornate.png` - Decorative divider line

### Custom Fonts (Optional)
- Crimson Text (or similar serif font)
- Could bundle custom medieval-style fonts

### Color Assets
Define in Assets.xcassets:
- ParchmentColor
- TealPrimary
- BurgundyAccent
- GoldSelection
- etc.

---

## Accessibility

### VoiceOver Support
- All interactive elements labeled
- Task status announced
- Priority level announced
- Progress percentages announced

### Keyboard Navigation
- Tab through tasks
- Space to select/toggle
- Arrow keys to navigate
- Enter to open details

### Color Contrast
- Ensure WCAG AA compliance
- Test with color blindness simulators
- Provide high-contrast mode option

### Dynamic Type
- Support system font scaling
- Maintain layout at larger sizes

---

## Implementation Notes

### SwiftUI Advantages
- Declarative syntax for complex layouts
- Native animations
- State management with @ObservedObject
- Preview canvas for rapid iteration

### Performance Considerations
- Lazy loading for large task lists
- Virtualized scrolling
- Debounced search/filter
- Efficient data binding

### Platform Integration
- Native macOS controls where appropriate
- Drag & drop support
- Context menus
- Toolbar items
- Keyboard shortcuts

---

## Future Enhancements

### Visual Enhancements
- Page turn animations between tabs
- Particle effects on task completion
- Animated ink effects for checkmarks
- More elaborate corner decorations

### Interactive Features
- Task templates with visual previews
- Drag-and-drop task organization
- Quick add with inline editing
- Batch operations with multi-select

### Customization
- Theme variants (different journal styles)
- Custom fonts selection
- Adjustable color schemes
- Layout density options

---

## Summary

The Quest Todo frontend design captures the essence of Trails series quest menus:

✨ **Authentic JRPG aesthetic** - Parchment, ornate borders, quest-style layout
⭐ **Priority system** - Visual star ratings (★★★★★)
📅 **Deadline indicators** - Color-coded urgency levels
✅ **Objective tracking** - Star-marked checklist with progress
📖 **Journal presentation** - Two-panel book-like layout
🎨 **Cohesive design** - Teal, gold, and parchment color scheme

The design balances aesthetic authenticity with modern usability, creating an engaging and productive task management experience.
